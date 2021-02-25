let fileInput = null;

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
        } catch (e) {
            openPopup('Error', e.message, [{
                text: 'Ok',
                onclick: closePopup
            }]);
            createNewFileInput();
            return;
        }

        console.log(object);

        location.replace(`/object/${object.id}`);
    });

    document.body.appendChild(fileInput);

    return fileInput;
}

window.addEventListener('load', function() {
    let createButton = document.getElementById('create-button');
    createNewFileInput();

    createButton.addEventListener('click', function() {
        fileInput.click();
    });
});

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
