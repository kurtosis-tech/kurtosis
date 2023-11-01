import { Textarea } from "@chakra-ui/react";
import { stringifyError } from "../../../../utils";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

export const JSONArgumentInput = (props: Omit<KurtosisArgumentTypeInputProps, "type">) => {
  const { register } = useEnclaveConfigurationFormContext();

  return (
    <Textarea
      {...register(props.name, {
        disabled: props.disabled,
        required: props.isRequired,
        value: "{}",
        validate: (value: string) => {
          try {
            JSON.parse(value);
          } catch (err: any) {
            return `This is not valid JSON. ${stringifyError(err)}`;
          }
        },
      })}
    />
  );
};
