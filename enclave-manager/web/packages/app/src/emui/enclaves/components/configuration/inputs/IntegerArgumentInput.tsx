import { Input } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

export const IntegerArgumentInput = (props: KurtosisArgumentTypeInputImplProps) => {
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
      placeholder={props.placeholder}
      width={props.width}
      size={props.size || "lg"}
      tabIndex={props.tabIndex}
    />
  );
};
