import { Input, InputProps } from "@chakra-ui/react";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

type StringArgumentInputProps = KurtosisArgumentTypeInputImplProps & InputProps;

export const StringArgumentInput = ({
  name,
  placeholder,
  isRequired,
  validate,
  disabled,
  width,
  size,
  tabIndex,
  ...inputProps
}: StringArgumentInputProps) => {
  const { register } = useEnclaveConfigurationFormContext();

  return (
    <Input
      {...register(name, { disabled: disabled, required: isRequired, validate: validate })}
      placeholder={placeholder}
      width={width}
      size={size || "lg"}
      tabIndex={tabIndex}
      {...inputProps}
    />
  );
};
