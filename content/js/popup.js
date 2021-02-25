(function() {
    let popupElement = null;
    let titleElement = null;
    let textElement = null;
    let buttonsPanel = null;

    function openPopup(title, text, buttons, options) {
        closePopup();

        popupElement = document.createElement('div');
        popupElement.id = 'popup';
        if (options && options.popupStyle) {
            for (let [key, value] of Object.entries(options.popupStyle)) {
                popupElement.style[key] = value;
            }
        }

        titleElement = document.createElement('div');
        titleElement.id = 'popup-title';
        titleElement.innerText = title ?? '';
        if (options && options.titleStyle) {
            for (let [key, value] of Object.entries(options.titleStyle)) {
                titleElement.style[key] = value;
            }
        }

        let separatorElement = document.createElement('div');
        separatorElement.id = 'popup-separator';

        textElement = document.createElement('div');
        textElement.id = 'popup-text';
        textElement.innerText = text ?? '';
        if (options && options.textStyle) {
            for (let [key, value] of Object.entries(options.textStyle)) {
                textStyle.style[key] = value;
            }
        }

        buttonsPanel = null;
        if (buttons) {
            for (let button of buttons) {
                let buttonElement = document.createElement('div');
                buttonElement.className = 'popup-button';
                buttonElement.innerText = button.text ?? '';
                buttonElement.addEventListener('click', button.onclick);

                if (!buttonsPanel) {
                    buttonsPanel = document.createElement('div');
                    buttonsPanel.id = 'popup-buttons-panel';
                }

                buttonsPanel.appendChild(buttonElement);
            }
        }

        popupElement.appendChild(titleElement);
        popupElement.appendChild(separatorElement);
        popupElement.appendChild(textElement);
        if (buttonsPanel) {
            popupElement.appendChild(buttonsPanel);
        }

        document.body.appendChild(popupElement);
    }

    function closePopup() {
        if (!popupElement) {
            return
        }

        document.body.removeChild(popupElement);
        popupElement = null;
    }

    function updatePopupText(text) {
        if (!popupElement) {
            return;
        }

        textElement.innerText = text ?? '';
    }

    function addPopupButton(button) {
        if (!popupElement) {
            return;
        }

        if (!buttonsPanel) {
            buttonsPanel = document.createElement('div');
            buttonsPanel.id = 'popup-buttons-panel';
            popupElement.appendChild(buttonsPanel)
        }

        let buttonElement = document.createElement('div');
        buttonElement.className = 'popup-button';
        buttonElement.innerText = button.text ?? '';
        buttonElement.addEventListener('click', button.onclick);

        buttonsPanel.appendChild(buttonElement);
    }

    window.openPopup = openPopup;
    window.closePopup = closePopup;
    window.updatePopupText = updatePopupText;
    window.addPopupButton = addPopupButton;
})();
