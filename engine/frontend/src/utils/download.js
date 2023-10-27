const DEFAULT_ELEMENT_NAME = "a"
const DEFAULT_FILE_TYPE = "text/plain"

export const saveTextAsFile = (text, fileName, options = {}) => {
    const elementName = options["elementName"] ? options["elementName"] : DEFAULT_ELEMENT_NAME
    const fileType = options["fileType"] ? options["fileType"] : DEFAULT_FILE_TYPE

    const blob = new Blob([text], {type: fileType});
    const downloadLink = document.createElement(elementName);
    downloadLink.download = fileName;
    downloadLink.innerHTML = "Download File";
    if (window.webkitURL) {
        // No need to add the download element to the DOM in Webkit.
        downloadLink.href = window.webkitURL.createObjectURL(blob);
    } else {
        downloadLink.href = window.URL.createObjectURL(blob);
        downloadLink.onclick = (event) => {
            if (event.target) {
                document.body.removeChild(event.target);
            }
        };
        downloadLink.style.display = "none";
        document.body.appendChild(downloadLink);
    }

    downloadLink.click();

    if (window.webkitURL) {
        window.webkitURL.revokeObjectURL(downloadLink.href);
    } else {
        window.URL.revokeObjectURL(downloadLink.href);
    }
  };
