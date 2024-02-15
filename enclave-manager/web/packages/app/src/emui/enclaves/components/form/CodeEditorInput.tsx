import { CodeEditor } from "kurtosis-ui-components";
import { Controller } from "react-hook-form";
import { FieldPath, FieldValues } from "react-hook-form/dist/types";
import { ControllerRenderProps } from "react-hook-form/dist/types/controller";

import { KurtosisFormInputProps } from "./types";

type CodeEditorInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  fileName: string;
};

export const CodeEditorInput = <DataModel extends object>(props: CodeEditorInputProps<DataModel>) => {
  return (
    <Controller
      render={({ field }) => <CodeEditorInputImpl field={field} fileName={props.fileName} />}
      name={props.name}
      defaultValue={"" as any}
      rules={{
        required: props.isRequired,
        validate: props.validate,
      }}
      disabled={props.disabled}
    />
  );
};

type CodeEditorImplProps<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
  field: ControllerRenderProps<TFieldValues, TName>;
  fileName: string;
};

const CodeEditorInputImpl = <
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
  field,
  fileName,
}: CodeEditorImplProps<TFieldValues, TName>) => {
  return <CodeEditor text={field.value} onTextChange={field.onChange} fileName={fileName} />;
};
