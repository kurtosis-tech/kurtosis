import { Input } from "@chakra-ui/react";
import { isDefined } from "../../../../utils";
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

          const propsValidation = props.validate ? props.validate(value) : undefined;
          if (isDefined(propsValidation)) {
            return propsValidation;
          }
        },
      })}
      width={props.width}
      size={props.size || "lg"}
    />
  );
};
