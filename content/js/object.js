let fileInput = null;
let elementWithFile = null;

function createNewFileInput() {
    if (fileInput) {
        document.body.removeChild(fileInput);
    }

    fileInput = document.createElement('input');
    fileInput.type = 'file';
    fileInput.id = 'file-input';
    fileInput.hidden = true;
    fileInput.addEventListener('change', async function(e) {
        let file = fileInput.files[0];

        let object;
        try {
            object = await createObject(file.name, file);
            document.body.removeChild(fileInput);
        } catch (e) {
            openPopup('Error', e.message, [{
                text: 'Ok',
                onclick: closePopup
            }]);
            createNewFileInput();
            return;
        }

        updatePopupText('File uploaded');
        location.replace(`/object/${object.id}`);
    });

    document.body.appendChild(fileInput);

    return fileInput;
}

window.addEventListener('load', async function() {
    let filenameElement = document.getElementById('filename');
    let filesizeElement = document.getElementById('filesize-value');
    let downloadsCountElement = document.getElementById('downloads-count-value');
    let iconImageElement = document.getElementById('icon-image');
    let downloadButton = document.getElementById('download-button');
    let createButton = document.getElementById('create-button');
    let copyLinkButton = document.getElementById('copy-link-button');
    createNewFileInput();

    let splittedPath = location.pathname.split('/');
    let objectId = splittedPath[splittedPath.length - 1];

    let object;
    try {
        object = await getObjectMetadata(objectId);
    } catch (e) {
        openPopup('Error', e.message, [{
            text: 'Ok',
            onclick: closePopup
        }]);
        return;
    }

    window.title = `Object ${object.id} | Storage`;

    filenameElement.innerText = object.filename;
    filesizeElement.innerText = humanReadableSize(object.size);
    downloadsCountElement.innerText = object.downloads.toString();

    iconImageElement.src = '/getIconByObjectType?type=' + object.type;
    iconImageElement.addEventListener('loal', function() {
        console.log('If I had a preloader, I must remove at this moment');
    });

    downloadButton.addEventListener('click', async function() {
        let file = await downloadObject(object);
        console.log(file);
    });

    createButton.addEventListener('click', function() {
        fileInput.click();
    });

    let returnTextTimeout = null;
    copyLinkButton.addEventListener('click', function() {
        if (returnTextTimeout) {
            clearTimeout(returnTextTimeout);
        }

        let temporaryButtonText = 'Copied!';

        try {
            navigator.clipboard.writeText(location.href);
        } catch (e) {
            temporaryButtonText = 'Can\'t copy :(';
        }

        let buttonTextElement = copyLinkButton.getElementsByClassName('button-text')[0];
        buttonTextElement.innerText = temporaryButtonText;

        returnTextTimeout = setTimeout(function() {
            buttonTextElement.innerText = 'Copy link';
        }, 1500);
    });
});

function humanReadableSize(size) {
    if (size < 1024 * 1024) {
        return (size / 1024).toFixed(2) + 'Kb';
    } else if (size < 1024 * 1024 * 1024) {
        return (size / 1024 / 1024).toFixed(2) + 'Mb';
    } else {
        return (size / 1024 / 1024 / 1024).toFixed(2) + 'Gb';
    }
}

async function getObjectMetadata(id) {
    let response = await fetch(`/getObjectMetadata?id=${id}`, {
        method: 'GET'
    });

    let data = await response.json()
    console.log(data);

    if (data.error) {
        throw new Error(data.error);
    }

    return data.object;
}

function createObject(filename, file) {
    return new Promise((resolve, reject) => {
        if (file.size > 1024 * 1024 * 1024) {
            reject(new Error('File too big'));
            return;
        }

        openPopup('Upload', 'Upload will be started soon');

        let xhr = new XMLHttpRequest();

        xhr.open('POST', '/createObject', true);

        xhr.setRequestHeader('x-file-name', filename);

        xhr.upload.onprogress = function(e) {
            let progress = (e.loaded / e.total * 100).toFixed() + '%';
            updatePopupText(`Uploaded: ${progress}`);
        };

        xhr.upload.onerror = function(e) {
            reject(new Error('Failed to upload'));
        }

        xhr.onload = function(e) {
            let response = JSON.parse(xhr.response);
            if (response.error) {
                reject(new Error(response.error));
            }

            resolve(response.object);
        }

        xhr.onerror = function(e) {
            reject(new Error('Failed to fetch response'));
        }

        xhr.send(file);
    });
}

function downloadObject(object) {
    if (elementWithFile) {
        openPopup('Download...', 'Download completed', [{
            text: 'Close',
            onclick: closePopup
        }]);
        elementWithFile.click();
        return;
    }

    openPopup('Download...', 'File will be downloaded soon');

    let xhr = new XMLHttpRequest();
    xhr.open('GET', `/downloadObject/${object.id}`, true);
    xhr.responseType = 'blob';

    xhr.onprogress = function(e) {
        let progress = (e.loaded / object.size * 100).toFixed(0) + '%';
        updatePopupText(`Progress: ${progress}`);
    };

    xhr.onload = function(e) {
        console.log(xhr.status);
        if (xhr.status == 200) {
            let reponse = xhr.response;
            console.log(reponse);

            let blob = xhr.response;
            let blobUrl = window.URL.createObjectURL(blob);

            elementWithFile = document.createElement('a');
            elementWithFile.href = blobUrl;
            elementWithFile.download = object.filename;
            elementWithFile.hidden = true;

            document.body.appendChild(elementWithFile);

            elementWithFile.click();

            updatePopupText('Download completed');

            addPopupButton({
                text: 'Close',
                onclick: closePopup
            });
        } else if (xhr.status == 404) {
            openPopup('Object Not Found', 'I am sorry :(', [{
                text: 'Ok',
                onclick: closePopup
            }]);
        } else if (xhr.status == 500) {
            openPopup('Server Internal Error', 'I am sorry :(', [{
                text: 'Ok',
                onclick: closePopup
            }]);
        }
    };

    xhr.onerror = function(e) {
        openPopup("Error", "Unknown error. May be lost Internet connection?", [{
            text: 'Ok',
            onclick: closePopup
        }]);
    };

    xhr.send();
}

