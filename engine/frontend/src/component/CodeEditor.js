import React, {useEffect, useRef, useState} from "react";
import {Box, Button, useClipboard} from "@chakra-ui/react";
import useWindowDimensions from "../utils/windowheight";
import Editor from "@monaco-editor/react";

export const CodeEditor = (
    dataCallback,
    readOnly = false,
    fullFileName = "json_field.json",
    languages = ["json"],
    defaultWidthPx = 500,
    defaultState = languages.includes("json") ? "{\n}" : "",
) => {
    // https://github.com/microsoft/monaco-editor/blob/main/webpack-plugin/README.md#options
    const [value, setValue] = useState(defaultState)
    const contentClipboard = useClipboard("");
    const monacoRef = useRef(null);
    const dimensions = useWindowDimensions();
    console.log("defaultState", defaultState)

    console.log(value)

    const getEditor = () => {
        if (!monacoRef.current) return null;
        return monacoRef.current.editor.getEditors()[0];
    }

    useEffect(() => {
        handleEditorChange(value)
    }, [])

    useEffect(() => {
        contentClipboard.setValue(value)
        // Resize view on content change
        updateWindowHeight();
    }, [value])

    // Resize view on window change
    useEffect(() => {
        if (getEditor()) {
            getEditor().layout({width: defaultWidthPx, height: getEditor().getContentHeight()});
            getEditor().layout()
        }
    }, [dimensions])

    const updateWindowHeight = () => {
        if (getEditor()) {
            const contentHeight = Math.min(1000, getEditor().getContentHeight());
            getEditor().layout({width: defaultWidthPx, height: contentHeight});
            getEditor().layout()
        }
    };

    const saveTextAsFile = (text, fileName) => {
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

    function handleEditorChange(value) {
        try {
            setValue(value)
            // const parsedJson = JSON.parse(value)
            // const jsonCleanedMinified = JSON.stringify(parsedJson)
            // dataCallback(jsonCleanedMinified)
            dataCallback(value)
        } catch (error) {
            // swallow
        }
    }

    function handleEditorDidMount(editor, monaco) {
        monacoRef.current = monaco;
        updateWindowHeight();
    }

    function handleDownload() {
        saveTextAsFile(value, fullFileName)
    }

    // TODO: We can use this to display error messages
    // function handleEditorValidation(markers) {
    //     // model markers
    //     // markers.forEach(marker => console.log('onValidate:', marker.message));
    // }

    return (
        <Box
            border="1px"
            borderColor='gray.700'
            borderRadius="7"
            margin={"1px"}
            padding={1}
        >
            <Editor
                margin={1}
                defaultLanguage="json"
                value={value}
                theme={"vs-dark"}
                onMount={handleEditorDidMount}
                onChange={handleEditorChange}
                // onValidate={handleEditorValidation}
                options={{
                    automaticLayout: true,
                    selectOnLineNumbers: true,
                    languages: languages,
                    readOnly: readOnly,
                    minimap: {
                        enabled: false
                    },
                    scrollBeyondLastLine: false
                }}
            />
            <Button
                margin={1}
                onClick={contentClipboard.onCopy}
            >
                {contentClipboard.hasCopied ? "Copied!" : "Copy"}
            </Button>
            <Button
                margin={1}
                onClick={handleDownload}

            >
                Download
            </Button>
        </Box>
    )
}
