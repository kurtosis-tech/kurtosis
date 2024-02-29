import { Select, SelectProps } from "@chakra-ui/react";

import { useFormContext } from "react-hook-form";
import { KurtosisFormInputProps } from "./types";

export type SelectOption = { value: string; display: string };

type SelectArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> &
  SelectProps & {
    options: SelectOption[];
  };

export const SelectArgumentInput = <DataModel extends object>({
  options,
  name,
  disabled,
  isRequired,
  validate,
  ...props
}: SelectArgumentInputProps<DataModel>) => {
  const { register } = useFormContext<DataModel>();

  return (
    <Select {...register(name, { required: isRequired, validate: validate })} isDisabled={disabled} {...props}>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.display}
        </option>
      ))}
    </Select>
  );
};
