import { Box } from "@chakra-ui/react";
import { Editor, OnChange, OnMount } from "@monaco-editor/react";
import { editor } from "monaco-editor";
import { useState } from "react";
import { isDefined } from "../utils";

type CodeEditorProps = {
  text: string;
  onTextChange?: (newText: string) => void;
  showLineNumbers?: boolean;
};

export const CodeEditor = ({ text, onTextChange, showLineNumbers }: CodeEditorProps) => {
  const isReadOnly = !isDefined(onTextChange);
  const [editor, setEditor] = useState<editor.IStandaloneCodeEditor>();

  const resizeEditorBasedOnContent = () => {
    if (isDefined(editor)) {
      // An initial layout call is needed, else getContentHeight is garbage
      editor.layout();
      const contentHeight = editor.getContentHeight();
      editor.layout({ width: 500, height: contentHeight });
      // Unclear why layout must be called twice, but seems to be necessary
      editor.layout();
    }
  };

  const handleMount: OnMount = (editor, monaco) => {
    setEditor(editor);
    monaco.editor.defineTheme("kurtosis-theme", {
      base: "vs-dark",
      inherit: true,
      rules: [],
      colors: {},
    });
    monaco.editor.setTheme("kurtosis-theme");
  };

  const handleChange: OnChange = (value, ev) => {
    if (isDefined(value) && onTextChange) {
      onTextChange(value);
      resizeEditorBasedOnContent();
    }
  };

  // Triggering this on every render seems to keep the editor correctly sized
  // it is unclear why this is the case.
  resizeEditorBasedOnContent();

  return (
    <Box width={"100%"} maxHeight={"1000px"}>
      <Editor
        onMount={handleMount}
        value={text}
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
        defaultLanguage={"json"}
        theme={"vs-dark"}
      />
    </Box>
  );
};
