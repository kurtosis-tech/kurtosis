import React, {useEffect, useRef, useState} from "react";
import {Box, Button, useClipboard} from "@chakra-ui/react";
import useWindowDimensions from "../utils/windowDimension";
import {saveTextAsFile} from "../utils/download";
import Editor from "@monaco-editor/react";

export const CodeEditor = (
    {
        uniqueId,
        dataCallback = (data) => {
        },
        readOnly = false,
        languages = ["json"],
        defaultWidthPx = 500,
        defaultState = languages.includes("json") ? "{\n}" : "",
        autoFormat = false,
        lineNumbers = false,
        showCopyButton = true,
        showDownloadButton = true,
        showFormatButton = true,
        buttonSizes = "sm",
        border = "1px",
        theme = "vs-dark",
    }
) => {
    // https://github.com/microsoft/monaco-editor/blob/main/webpack-plugin/README.md#options
    const monacoRef = useRef(null);
    const dimensions = useWindowDimensions();
    const [formatCode, setFormatCode] = useState(false)
    const [monacoReadOnlySettingToggle, setMonacoReadOnlySettingToggle] = useState(false)
    // TODO: This may need to be by Model in the future (i.e. state is indexed by unique id):
    const contentClipboard = useClipboard("");
    const originalReadOnlySetting = useRef(readOnly)
    const [readOnlySetting, setReadOnlySetting] = useState(readOnly)

    // TODO: This could lead to bugs in the future. This number depends on the version of Monaco! Use actual enum instead.
    const monacoReadOnlyEnumId = 86;

    // TODO: Add a promise to eventually return the editor, and make all callers process async
    const getEditor = () => {
        if (!monacoRef.current) return null;
        return findEditorByModelUri(uniqueId);
    }

    const findEditorByModelUri = (modelUri) => {
        return monacoRef.current.editor.getEditors().find((model) => {
            return model.getModel().uri.toString() === `file:///${modelUri}`
        })
    }

    const getValue = () => {
        return getEditor()?.getModel().getValue();
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
                setMonacoReadOnlySettingToggle(!monacoReadOnlySettingToggle)
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
                        })
                        .catch((e) => {
                            console.error("An error happened", e)
                        });
                }
            }
        }
    }, [formatCode, monacoReadOnlySettingToggle])

    // Resize view on window change
    useEffect(() => updateWindowHeightBasedOnContent(), [dimensions])

    useEffect(() => handleEditorChange(defaultState), [defaultState])

    const updateWindowHeightBasedOnContent = () => {
        if (getEditor()) {
            const contentHeight = Math.min(750, getEditor().getContentHeight());
            getEditor().layout({width: defaultWidthPx, height: contentHeight});
            getEditor().layout()
        }
    };

    const handleEditorChange = (value) => {
        updateWindowHeightBasedOnContent();
        dataCallback(value);
        contentClipboard.setValue(value);
    }

    const handleEditorDidMount = (editor, monaco) => {
        monacoRef.current = monaco;
        updateWindowHeightBasedOnContent();
        attachOptionsChangeListener()
        if (autoFormat) handleCodeFormat();
        defineAndSetTheme('kurtosis-theme', {
            base: theme,
            inherit: true,
            rules: [],
            colors: {
                'editor.background': '#171923',
            },
        });
    }

    const handleDownload = () => {
        saveTextAsFile(getValue(), uniqueId)
    }

    const handleCodeFormat = () => {
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

    const copyToClipboard = () => {
        contentClipboard.setValue(defaultState);
        contentClipboard.onCopy();
    }

    updateWindowHeightBasedOnContent();

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
                value={defaultState}
                theme={theme}
                path={uniqueId}
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
                        verticalScrollbarSize: scrollbarOption(readOnlySetting),
                        alwaysConsumeMouseWheel: !readOnly, // We want the original read only setting
                    },
                    renderLineHighlight: highlightLineOption(readOnlySetting),
                }}
            />
            <Box
                marginTop={1}
            >
                {showCopyButton && (
                    <Button
                        margin={1}
                        onClick={copyToClipboard}
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
