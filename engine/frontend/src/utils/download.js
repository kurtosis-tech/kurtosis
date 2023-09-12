export const saveTextAsFile = (text, fileName) => {
    const blob = new Blob([text], {type: "text/plain"});
    const downloadLink = document.createElement("a");
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