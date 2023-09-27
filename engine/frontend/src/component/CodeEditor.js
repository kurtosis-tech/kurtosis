import React, {useEffect, useMemo, useRef, useState} from "react";
import {Box, Button, useClipboard} from "@chakra-ui/react";
import useWindowDimensions from "../utils/windowDimension";
import {saveTextAsFile} from "../utils/download";
import Editor from "@monaco-editor/react";

export const CodeEditor = (
    dataCallback,
    readOnly = false,
    fullFileName = "json_field.json",
    languages = ["json"],
    defaultWidthPx = 500,
    defaultState = languages.includes("json") ? "{\n}" : "",
    autoFormat = false,
    lineNumbers = false,
    id = 0,
    showCopyButton = true,
    showDownloadButton = true,
    showFormatButton = true,
    buttonSizes = "sm",
    border = "1px"
) => {
    // https://github.com/microsoft/monaco-editor/blob/main/webpack-plugin/README.md#options
    const [value, setValue] = useState(defaultState)
    const contentClipboard = useClipboard("");
    const monacoRef = useRef(null);
    const dimensions = useWindowDimensions();
    const originalReadOnlySetting = useRef(readOnly)
    const [readOnlySetting, setReadOnlySetting] = useState(readOnly)
    const [formatCode, setFormatCode] = useState(false)
    const [monacoReadOnlySettingHasChanged, setMonacoReadOnlySettingHasChangedHasChanged] = useState(false)
    // TODO: This could lead to bugs in the future:
    //  This number depends on the version of Monaco! Use actual enum instead.
    const monacoReadOnlyEnumId = 86;

    // TODO: Add a promise to getEditor()
    const getEditor = () => {
        if (!monacoRef.current) return null;
        return monacoRef.current.editor.getEditors()[id];
    }

    const defineAndSetTheme = (name, data) => {
        monacoRef.current.editor.defineTheme(name, data);
        monacoRef.current.editor.setTheme(name);
    }

    const isEditorReadOnly = () => {
        try {
            return getEditor().getOption(monacoReadOnlyEnumId)
        } catch (e) {
            return undefined
        }
    }

    function attachOptionsChangeListener() {
        getEditor().onDidChangeConfiguration((event) => {
            if (event.hasChanged(monacoReadOnlyEnumId)) {
                setMonacoReadOnlySettingHasChangedHasChanged(true)
            }
        });
    }

    useEffect(() => {
        if (getEditor()) {
            getEditor().updateOptions({
                readOnly: readOnlySetting,
            })
        }
    }, [readOnlySetting])

    useEffect(() => {
        if (formatCode) {
            if (isEditorReadOnly()) {
                setReadOnlySetting(false)
            } else {
                if (getEditor()) {
                    getEditor()
                        .getAction('editor.action.formatDocument')
                        .run()
                        .then(() => {
                            setReadOnlySetting(originalReadOnlySetting.current)
                            setFormatCode(false)
                        });
                }
            }
        }
    }, [formatCode, monacoReadOnlySettingHasChanged])

    // Start by manually setting the content of the editor. From hereafter user interaction will update it:
    useEffect(() => {
        handleEditorChange(value)
    }, [])

    useEffect(() => {
        contentClipboard.setValue(value)
        // Resize view on content change
        updateWindowHeightBasedOnContent();
    }, [value])

    // Resize view on window change
    useEffect(() => {
        if (getEditor()) {
            getEditor().layout({width: defaultWidthPx, height: getEditor().getContentHeight()});
            getEditor().layout()
        }
    }, [dimensions])

    const updateWindowHeightBasedOnContent = () => {
        if (getEditor()) {
            const contentHeight = Math.min(1000, getEditor().getContentHeight());
            getEditor().layout({width: defaultWidthPx, height: contentHeight});
            getEditor().layout()
        }
    };

    function handleEditorChange(value) {
        setValue(value)
        dataCallback(value)
    }

    function handleEditorDidMount(editor, monaco) {
        monacoRef.current = monaco;
        updateWindowHeightBasedOnContent();
        attachOptionsChangeListener()
        if (autoFormat) handleCodeFormat();
        defineAndSetTheme('kurtosisTheme', {
            base: 'vs-dark',
            inherit: true,
            rules: [],
            colors: {
                'editor.background': '#171923',
            },
        });
    }

    function handleDownload() {
        saveTextAsFile(value, fullFileName)
    }

    function handleCodeFormat() {
        setFormatCode(true)
    }

    // TODO: We can use this to display error messages
    // function handleEditorValidation(markers) {
    //     // model markers
    //     // markers.forEach(marker => console.log('onValidate:', marker.message));
    // }

    const isNotFormattable = () => {
        return !languages.includes("json")
    }

    const highlightLineOption = (readOnly) => {
        if (readOnly) {
            return "none"
        } else {
            return "line"
        }
    }

    const scrollbarOption = (readOnly) => {
        if (readOnly) {
            return 0
        } else {
            return 10
        }
    }

    return (
        <Box
            border={border}
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
                    automaticLayout: false, // if this is `true` a ResizeObserver is installed. This causes issues with us managing the container size outside.
                    selectOnLineNumbers: lineNumbers,
                    lineNumbers: lineNumbers,
                    languages: languages,
                    readOnly: readOnlySetting,
                    minimap: {
                        enabled: false
                    },
                    scrollBeyondLastLine: false,
                    scrollbar: {
                        verticalScrollbarSize: scrollbarOption(readOnly),
                    },
                    renderLineHighlight: highlightLineOption(readOnly),
                }}
            />
            <Box
                marginTop={1}
            >
                {showCopyButton && (
                    <Button
                        margin={1}
                        onClick={contentClipboard.onCopy}
                        size={buttonSizes}
                    >
                        {contentClipboard.hasCopied ? "Copied!" : "Copy"}
                    </Button>
                )}
                {showDownloadButton && (
                    <Button
                        margin={1}
                        onClick={handleDownload}
                        size={buttonSizes}
                    >
                        Download
                    </Button>
                )}
                {showFormatButton && (
                    <Button
                        margin={1}
                        onClick={handleCodeFormat}
                        isDisabled={isNotFormattable()}
                        size={buttonSizes}
                    >
                        Format
                    </Button>
                )}
            </Box>
        </Box>
    )
}
