import { Radio, RadioGroup, Stack } from "@chakra-ui/react";

import { useFormContext } from "react-hook-form";
import { KurtosisFormInputProps } from "./types";

type OptionsArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  options: string[];
};

export const OptionsArgumentInput = <DataModel extends object>({
  options,
  ...props
}: OptionsArgumentInputProps<DataModel>) => {
  const { register, getValues } = useFormContext<DataModel>();

  const currentDefault = getValues(props.name);

  return (
    <RadioGroup defaultValue={currentDefault}>
      <Stack direction={"row"}>
        {options.map((option) => (
          <Radio
            key={option}
            {...register(props.name, {
              disabled: props.disabled,
              required: props.isRequired,
              validate: props.validate,
            })}
            value={option}
          >
            {option}
          </Radio>
        ))}
      </Stack>
    </RadioGroup>
  );
};
