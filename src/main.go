package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/don2quixote/ninja"
)

func init() {
	readConfig()
	db, err := openSqlConnection()
	if err != nil {
		fmt.Println("Eror connecting database:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	err = db.initDatabase()
	if err != nil {
		fmt.Println("Error initializing database:", err.Error())
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	router := ninja.CreateRouter(100, 100)
	// router.SetMiddlewire("", ninja.ThroughMiddlewire(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println(r.URL.Path)
	// }))

	static := http.FileServer(http.Dir("content"))
	router.Handle("/object/", http.StripPrefix("/object", http.HandlerFunc(handleObjectRequest))).Methods("GET")
	router.HandleFunc("/getObjectMetadata", handleGetObjectMetadata).Methods("GET")
	router.HandleFunc("/getIconByObjectType", handleGetIconByObjectType).Methods("GET")
	router.HandleFunc("/createObject", handleCreateObject).Methods("POST")
	router.HandleFunc("/downloadObject/", handleDownloadObject).Methods("GET")
	router.Handle("/s/", http.StripPrefix("/s", static)).Methods("GET")
	router.HandleFunc("/", handleRootRequest).Methods("GET")

	err := http.ListenAndServe(":"+config.Port, router)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func handleNotFound(res http.ResponseWriter, req *http.Request) {
	file404, err := ioutil.ReadFile("content/html/404.html")
	if err != nil {
		res.Write([]byte("404 NOT FOUND"))
		return
	}

	res.Write(file404)
}

func handleRootRequest(res http.ResponseWriter, req *http.Request) {
	notFound := false
	switch req.URL.Path {
	case "/":
		file, err := ioutil.ReadFile("content/html/index.html")
		if err == nil {
			res.Write(file)
			break
		}
		notFound = true
		fallthrough
	default:
		if !notFound {
			file, err := ioutil.ReadFile("content/html" + req.URL.Path + ".html")
			if err == nil {
				res.Write(file)
				break
			}
		}

		handleNotFound(res, req)
	}
}

func handleCreateObject(res http.ResponseWriter, req *http.Request) {
	var responseStruct struct {
		Success bool    `json:"success"`
		Object  *object `json:"object,omitempty"`
		Error   string  `json:"error,omitempty"`
	}

	contentLengthHeader := req.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		responseStruct.Success = false
		responseStruct.Error = "Invalid Content-Length header"

		sendJSON(res, responseStruct)

		return
	}

	if contentLength > 1024*1024*1024 {
		responseStruct.Success = false
		responseStruct.Error = "Too large file"

		sendJSON(res, responseStruct)

		return
	} else if contentLength == 0 {
		responseStruct.Success = false
		responseStruct.Error = "Can't add object with 0 size"

		sendJSON(res, responseStruct)

		return
	}

	filename := req.Header.Get("x-file-name")
	if filename == "" {
		responseStruct.Success = false
		responseStruct.Error = "Cannot add object with empty name"

		sendJSON(res, responseStruct)

		return
	} else if len(filename) > 64 {
		responseStruct.Success = false
		responseStruct.Error = "Filename too long"

		sendJSON(res, responseStruct)

		return
	}

	fileExtension := ""
	splittedFilename := strings.Split(filename, ".")
	if len(splittedFilename) > 1 {
		fileExtension = splittedFilename[len(splittedFilename)-1]
	}

	db, err := openSqlConnection()
	if err != nil {
		println(err.Error())
		responseStruct.Success = false
		responseStruct.Error = "Internal Error"

		sendJSON(res, responseStruct)

		return
	}
	defer db.Close()

	objectIdBytes := make([]byte, 8)
	rand.Read(objectIdBytes)
	objectId := hex.EncodeToString(objectIdBytes)

	err = os.MkdirAll("objects/"+objectId, 0755)
	if err != nil {
		responseStruct.Success = false
		responseStruct.Error = "Internal Error"

		sendJSON(res, responseStruct)

		return
	}

	file, err := os.Create("objects/" + objectId + "/" + filename)
	if err != nil {
		responseStruct.Success = false
		responseStruct.Error = "Internal Error"

		sendJSON(res, responseStruct)

		return
	}

	var fileBegining []byte
	buffer := make([]byte, 4096)
	totalBytesRead := 0
	for {
		bytesRead, _ := req.Body.Read(buffer)
		if bytesRead != 4096 {
			fmt.Println("Bytes read != 4096. Its =", bytesRead)
		}

		if bytesRead == 0 {
			break
		}

		totalBytesRead += bytesRead
		if totalBytesRead > contentLength {
			responseStruct.Success = false
			responseStruct.Error = "Incorrect Content-Length header"

			sendJSON(res, responseStruct)

			file.Close()
			os.RemoveAll("objects/" + objectId)

			req.Body.Close()

			return
		}

		if fileBegining == nil {
			if bytesRead > 512 {
				fileBegining = append(fileBegining, buffer[0:512]...)
			} else {
				fileBegining = append(fileBegining, buffer[0:bytesRead]...)
			}
		}

		file.Write(buffer[0:bytesRead])
	}
	file.Close()
	req.Body.Close()

	if totalBytesRead != contentLength {
		responseStruct.Success = false
		responseStruct.Error = "Incorrect Content-Length header"

		sendJSON(res, responseStruct)

		os.RemoveAll("objects/" + objectId)

		return
	}

	contentType := http.DetectContentType(fileBegining)

	objType := typeBin
	if strings.Contains(contentType, "application/zip") {
		if fileExtension == "docx" || fileExtension == "doc" {
			objType = typeDoc
		} else if fileExtension == "jar" {
			objType = typeBin
		} else {
			objType = typeZip
		}
	} else if strings.Contains(contentType, "application/pdf") {
		objType = typePdf
	} else if strings.Contains(contentType, "text") {
		if fileExtension == "txt" {
			objType = typeTxt
		} else {
			objType = typeCode
		}
	} else if strings.Contains(contentType, "image") {
		objType = typeImage
	} else if strings.Contains(contentType, "video") {
		objType = typeVideo
	} else if strings.Contains(contentType, "audio") {
		objType = typeAudio
	}

	db.addObject(objectId, filename, contentLength, objType)

	responseStruct.Success = true
	responseStruct.Object = &object{
		Id:       objectId,
		Filename: filename,
		Size:     contentLength,
		Type:     objType,
	}

	sendJSON(res, responseStruct)
}

func handleObjectRequest(res http.ResponseWriter, req *http.Request) {
	splittedPath := strings.Split(req.URL.Path, "/")
	objectId := splittedPath[1]

	db, err := openSqlConnection()
	if err != nil {
		handleNotFound(res, req)
		return
	}
	defer db.Close()

	_, err = db.getObject(objectId)
	if err != nil {
		handleNotFound(res, req)
		return
	}

	file, err := ioutil.ReadFile("content/html/object.html")
	if err != nil {
		handleNotFound(res, req)
		return
	}

	res.Write(file)
}

func handleGetObjectMetadata(res http.ResponseWriter, req *http.Request) {
	var responseStruct struct {
		Success bool    `json:"success"`
		Object  *object `json:"object,omitempty"`
		Error   string  `json:"error,omitempty"`
	}

	objectId := req.URL.Query().Get("id")
	if objectId == "" {
		responseStruct.Success = false
		responseStruct.Error = "No id parameter"

		sendJSON(res, responseStruct)

		return
	}

	db, err := openSqlConnection()
	if err != nil {
		responseStruct.Success = false
		responseStruct.Error = "Internal Error"

		sendJSON(res, responseStruct)

		return
	}
	defer db.Close()

	object, err := db.getObject(objectId)
	if err != nil {
		responseStruct.Success = false
		responseStruct.Error = "Object not found"

		sendJSON(res, responseStruct)

		return
	}

	responseStruct.Success = true
	responseStruct.Object = &object

	sendJSON(res, responseStruct)
}

func handleGetIconByObjectType(res http.ResponseWriter, req *http.Request) {
	objType := objectType(req.URL.Query().Get("type"))
	switch objType {
	case typeZip:
		http.Redirect(res, req, "/s/assets/icons/zip-icon.svg", http.StatusMovedPermanently)
	case typeDoc:
		http.Redirect(res, req, "/s/assets/icons/doc-icon.svg", http.StatusMovedPermanently)
	case typeCode:
		http.Redirect(res, req, "/s/assets/icons/code-icon.svg", http.StatusMovedPermanently)
	case typeBin:
		http.Redirect(res, req, "/s/assets/icons/bin-icon.svg", http.StatusMovedPermanently)
	case typePdf:
		http.Redirect(res, req, "/s/assets/icons/pdf-icon.svg", http.StatusMovedPermanently)
	case typeImage:
		http.Redirect(res, req, "/s/assets/icons/image-icon.svg", http.StatusMovedPermanently)
	case typeAudio:
		http.Redirect(res, req, "/s/assets/icons/audio-icon.svg", http.StatusMovedPermanently)
	case typeVideo:
		http.Redirect(res, req, "/s/assets/icons/video-icon.svg", http.StatusMovedPermanently)
	case typeTxt:
		http.Redirect(res, req, "/s/assets/icons/txt-icon.svg", http.StatusMovedPermanently)
	default:
		http.Redirect(res, req, "/s/assets/icons/bin-icon.svg", http.StatusMovedPermanently)
	}
}

func handleDownloadObject(res http.ResponseWriter, req *http.Request) {
	splittedPath := strings.Split(req.URL.Path, "/")
	objectId := splittedPath[len(splittedPath)-1]

	db, err := openSqlConnection()
	if err != nil {
		println(err.Error())
		http.Error(res, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	object, err := db.getObject(objectId)
	if err != nil {
		println(err.Error())
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	file, err := os.Open("objects/" + objectId + "/" + object.Filename)
	if err != nil {
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	io.Copy(res, file)
}

func sendJSON(res http.ResponseWriter, data interface{}) {
	encoder := json.NewEncoder(res)
	encoder.Encode(data)
}
