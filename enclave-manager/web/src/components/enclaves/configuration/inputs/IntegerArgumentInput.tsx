import { Input } from "@chakra-ui/react";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

export const IntegerArgumentInput = (props: Omit<KurtosisArgumentTypeInputProps, "type">) => {
  const { register } = useEnclaveConfigurationFormContext();

  return (
    <Input
      {...register(props.name, {
        disabled: props.disabled,
        required: props.isRequired,
        validate: (value: number) => {
          if (isNaN(value)) {
            return "This value should be an integer";
          }
        },
      })}
    />
  );
};
