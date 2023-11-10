import { Controller } from "react-hook-form";
import { isDefined, stringifyError } from "../../../../utils";
import { CodeEditor } from "../../../CodeEditor";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

export const JSONArgumentInput = (props: Omit<KurtosisArgumentTypeInputProps, "type">) => {
  const { control } = useEnclaveConfigurationFormContext();

  return (
    <Controller
      render={({ field }) => <CodeEditor text={field.value} onTextChange={field.onChange} />}
      name={props.name}
      defaultValue={"{}"}
      rules={{
        required: props.isRequired,
        validate: (value: string) => {
          try {
            JSON.parse(value);
          } catch (err: any) {
            return `This is not valid JSON. ${stringifyError(err)}`;
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
