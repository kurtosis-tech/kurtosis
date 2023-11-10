import { Box } from "@chakra-ui/react";
import { Editor, OnChange, OnMount } from "@monaco-editor/react";
import { editor } from "monaco-editor";
import { useEffect, useState } from "react";
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
      const contentHeight = Math.min(750, editor.getContentHeight() || 10);
      editor.layout({ width: 500, height: contentHeight });
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

  useEffect(() => resizeEditorBasedOnContent(), [editor]);

  const handleChange: OnChange = (value, ev) => {
    if (isDefined(value) && onTextChange) {
      onTextChange(value);
      resizeEditorBasedOnContent();
    }
  };

  return (
    <Box width={"100%"} minHeight={`${editor?.getLayoutInfo().height || 10}px`}>
      <Editor
        onMount={handleMount}
        value={text}
        onChange={handleChange}
        options={{
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
        height={"100%"}
      />
    </Box>
  );
};
