package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
}

type objectType string

const (
	typeZip   objectType = "zip"
	typeDoc   objectType = "doc"
	typeCode  objectType = "code"
	typeBin   objectType = "bin"
	typePdf   objectType = "pdf"
	typeImage objectType = "image"
	typeAudio objectType = "audio"
	typeVideo objectType = "video"
	typeTxt   objectType = "txt"
)

type object struct {
	Id       string     `json:"id"`
	Filename string     `json:"filename"`
	Size     int        `json:"size"`
	Type     objectType `json:"type"`
}

func openSqlConnection() (*DB, error) {
	sourceString := config.DatabaseUser + ":" + config.DatabasePass + "@tcp(" + config.DatabaseHost + ":" + config.DatabasePort + ")/" + config.DatabaseName
	db, err := sql.Open("mysql", sourceString)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) initDatabase() error {
	query := "CREATE TABLE IF NOT EXISTS objects ( " +
		"id VARCHAR(16) PRIMARY KEY, " +
		"name VARCHAR(64) DEFAULT 'object' NOT NULL, " +
		"size INT NOT NULL, " +
		"type ENUM('zip', 'doc', 'code', 'bin', 'pdf', 'image', 'audio', 'video', 'txt') DEFAULT 'bin' NOT NULL " +
		")"
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) addObject(id string, name string, size int, t objectType) error {
	query := "INSERT INTO objects " +
		"(id, name, size, type) " +
		"VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, id, name, size, t)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) getObject(objectId string) (object, error) {
	query := "SELECT name, size, type " +
		"FROM objects " +
		"WHERE id = ? " +
		"LIMIT 1"
	row := db.QueryRow(query, objectId)

	var (
		filename string
		size     int
		objType  objectType
	)
	err := row.Scan(&filename, &size, &objType)
	if err != nil {
		return object{}, err
	}

	return object{
		Id:       objectId,
		Filename: filename,
		Size:     size,
		Type:     objType,
	}, nil
}

func (db *DB) removeObject(objectId string) error {
	query := "DELETE FROM objects " +
		"WHERE id = ?"
	_, err := db.Exec(query, objectId)
	if err != nil {
		return err
	}

	return nil
}
