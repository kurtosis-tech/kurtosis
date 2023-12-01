import { Box } from "@chakra-ui/react";
import { Editor, Monaco, OnChange, OnMount } from "@monaco-editor/react";
import { type editor as monacoEditor } from "monaco-editor";
import { forwardRef, useCallback, useEffect, useImperativeHandle, useState } from "react";
import { assertDefined, isDefined } from "../utils";

type CodeEditorProps = {
  text: string;
  fileName?: string;
  onTextChange?: (newText: string) => void;
  showLineNumbers?: boolean;
};

export type CodeEditorImperativeAttributes = {
  formatCode: () => Promise<void>;
  setText: (text: string) => void;
  setLanguage: (language: string) => void;
};

export const CodeEditor = forwardRef<CodeEditorImperativeAttributes, CodeEditorProps>(
  ({ text, fileName, onTextChange, showLineNumbers }, ref) => {
    const isReadOnly = !isDefined(onTextChange);
    const [monaco, setMonaco] = useState<Monaco>();
    const [editor, setEditor] = useState<monacoEditor.IStandaloneCodeEditor>();

    const resizeEditorBasedOnContent = useCallback(() => {
      if (isDefined(editor)) {
        // An initial layout call is needed, else getContentHeight is garbage
        editor.layout();
        const contentHeight = editor.getContentHeight();
        editor.layout({ width: editor.getContentWidth(), height: contentHeight });
        // Unclear why layout must be called twice, but seems to be necessary
        editor.layout();
      }
    }, [editor]);

    const handleMount: OnMount = (editor, monaco) => {
      setMonaco(monaco);
      setEditor(editor);
      const colors: monacoEditor.IColors = {};
      if (isReadOnly) {
        colors["editor.background"] = "#111111";
      }
      monaco.editor.defineTheme("kurtosis-theme", {
        base: "vs-dark",
        inherit: true,
        rules: [],
        colors,
      });
      monaco.editor.setTheme("kurtosis-theme");
    };

    const handleChange: OnChange = (value, ev) => {
      if (isDefined(value) && onTextChange) {
        onTextChange(value);
        resizeEditorBasedOnContent();
      }
    };

    useImperativeHandle(
      ref,
      () => ({
        formatCode: async () => {
          if (!isDefined(editor)) {
            // do nothing
            return;
          }
          if (isReadOnly) {
            return new Promise((resolve) => {
              const listenerDisposer = editor.onDidChangeConfiguration((event) => {
                if (event.hasChanged(89 /* ID of the readonly option */)) {
                  const formatAction = editor.getAction("editor.action.formatDocument");
                  assertDefined(formatAction, `Format action is not defined`);
                  formatAction.run().then(() => {
                    listenerDisposer.dispose();
                    editor.updateOptions({
                      readOnly: isReadOnly,
                    });
                    resizeEditorBasedOnContent();
                    resolve();
                  });
                }
              });
              editor.updateOptions({
                readOnly: false,
              });
            });
          } else {
            const formatAction = editor.getAction("editor.action.formatDocument");
            console.log(editor.getModel()?.getLanguageId());
            assertDefined(formatAction, `Format action is not defined`);
            return formatAction.run();
          }
        },
        setText: (text: string) => {
          if (!isDefined(editor)) {
            return;
          }
          editor.setValue(text);
        },
        setLanguage: (language: string) => {
          if (!isDefined(editor) || !isDefined(monaco)) {
            return;
          }
          const model = editor.getModel();
          if (!isDefined(model)) {
            return;
          }
          monaco.editor.setModelLanguage(model, language);
        },
      }),
      [isReadOnly, editor, monaco, resizeEditorBasedOnContent],
    );

    useEffect(() => {
      // Triggered as the text can change without internal editing. (ie if the
      // controlled prop changes)
      resizeEditorBasedOnContent();
    }, [text, resizeEditorBasedOnContent]);

    // Triggering this on every render seems to keep the editor correctly sized
    // it is unclear why this is the case.
    resizeEditorBasedOnContent();

    return (
      <Box width={"100%"}>
        <Editor
          onMount={handleMount}
          value={text}
          path={fileName}
          onChange={handleChange}
          options={{
            automaticLayout: false, // if this is `true` a ResizeObserver is installed. This causes issues with us managing the container size outside.
            readOnly: isReadOnly,
            lineNumbers: showLineNumbers || (!isDefined(showLineNumbers) && !isReadOnly) ? "on" : "off",
            minimap: { enabled: false },
            wordWrap: "on",
            wrappingStrategy: "advanced",
            scrollBeyondLastLine: false,
            renderLineHighlight: isReadOnly ? "none" : "line",
            selectionHighlight: !isReadOnly,
            occurrencesHighlight: !isReadOnly,
            overviewRulerLanes: isReadOnly ? 0 : 3,
            scrollbar: {
              alwaysConsumeMouseWheel: false,
            },
          }}
          defaultLanguage={!isDefined(fileName) ? "json" : undefined}
          theme={"vs-dark"}
        />
      </Box>
    );
  },
);
