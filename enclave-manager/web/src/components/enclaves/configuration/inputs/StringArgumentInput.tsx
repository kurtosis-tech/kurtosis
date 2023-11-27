import { Input } from "@chakra-ui/react";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

export const StringArgumentInput = (props: KurtosisArgumentTypeInputImplProps) => {
  const { register } = useEnclaveConfigurationFormContext();

  return (
    <Input
      {...register(props.name, { disabled: props.disabled, required: props.isRequired, validate: props.validate })}
      placeholder={props.placeholder}
      width={props.width}
      size={props.size || "lg"}
      tabIndex={props.tabIndex}
    />
  );
};
