import { Input, InputProps } from "@chakra-ui/react";

import { useFormContext } from "react-hook-form";
import { KurtosisFormInputProps } from "./types";

type StringArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & InputProps;

export const StringArgumentInput = <DataModel extends object>({
  name,
  placeholder,
  isRequired,
  validate,
  disabled,
  width,
  size,
  tabIndex,
  ...inputProps
}: StringArgumentInputProps<DataModel>) => {
  const { register } = useFormContext<DataModel>();

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
