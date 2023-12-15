import { Button, ButtonGroup, Flex } from "@chakra-ui/react";
import {
  CodeEditor,
  CodeEditorImperativeAttributes,
  FormatButton,
  isDefined,
  stringifyError,
} from "kurtosis-ui-components";
import { useEffect, useMemo, useRef, useState } from "react";
import { Controller } from "react-hook-form";
import { FieldPath, FieldValues } from "react-hook-form/dist/types";
import { ControllerRenderProps } from "react-hook-form/dist/types/controller";
import { FiCode } from "react-icons/fi";
import YAML from "yaml";
import { KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

export const JSONArgumentInput = (props: KurtosisArgumentTypeInputImplProps) => {
  return (
    <Controller
      render={({ field }) => <JsonAndYamlCodeEditor field={field} />}
      name={props.name}
      defaultValue={"{}"}
      rules={{
        required: props.isRequired,
        validate: (value: string) => {
          try {
            YAML.parse(value);
          } catch (err: any) {
            return `This is not valid JSON/YAML. ${stringifyError(err)}`;
          }

          const propsValidation = props.validate ? props.validate(value) : undefined;
          if (isDefined(propsValidation)) {
            return propsValidation;
          }
        },
      }}
      disabled={props.disabled}
    />
  );
};

const JsonAndYamlCodeEditor = <
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
  field,
}: {
  field: ControllerRenderProps<TFieldValues, TName>;
}) => {
  const [isWorking, setIsWorking] = useState(false);
  const codeEditorRef = useRef<CodeEditorImperativeAttributes>(null);
  const isProbablyJson = useMemo(() => isDefined(field.value.match(/^\s*[[{"]/g)), [field.value]);

  const handleFormatClick = async () => {
    if (isDefined(codeEditorRef.current)) {
      setIsWorking(true);
      codeEditorRef.current.setLanguage(isProbablyJson ? "json" : "yaml");
      await codeEditorRef.current.formatCode();
      setIsWorking(false);
    }
  };

  const handleConvertClick = () => {
    if (!isDefined(codeEditorRef.current)) {
      return;
    }
    try {
      if (isProbablyJson) {
        const newText = YAML.stringify(JSON.parse(field.value));
        codeEditorRef.current.setText(newText);
        codeEditorRef.current.setLanguage("yaml");
      } else {
        const newText = JSON.stringify(YAML.parse(field.value), undefined, 4);
        codeEditorRef.current.setText(newText);
        codeEditorRef.current.setLanguage("json");
      }
    } catch (err: any) {
      console.error(err);
    }
  };

  useEffect(() => {
    codeEditorRef.current?.setLanguage(isProbablyJson ? "json" : "yaml");
  }, [isProbablyJson]);

  return (
    <Flex flexDirection={"column"} gap={"10px"}>
      <ButtonGroup>
        <FormatButton size="xs" onClick={handleFormatClick} isLoading={isWorking} />
        <Button size={"xs"} colorScheme={"darkBlue"} onClick={handleConvertClick} leftIcon={<FiCode />}>
          Switch to {isProbablyJson ? "YAML" : "JSON"}
        </Button>
      </ButtonGroup>
      <CodeEditor ref={codeEditorRef} text={field.value} onTextChange={field.onChange} />
    </Flex>
  );
};
