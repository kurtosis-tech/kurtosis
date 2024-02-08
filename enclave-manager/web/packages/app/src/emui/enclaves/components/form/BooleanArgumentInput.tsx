import { Radio, RadioGroup, Stack, Switch } from "@chakra-ui/react";

import { useFormContext } from "react-hook-form";
import { KurtosisFormInputProps } from "./types";

type BooleanArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  inputType?: "radio" | "switch";
};

export const BooleanArgumentInput = <DataModel extends object>({
  inputType,
  ...props
}: BooleanArgumentInputProps<DataModel>) => {
  const { register, getValues } = useFormContext<DataModel>();

  const currentDefault = getValues(props.name);

  if (inputType === "switch") {
    return (
      <Switch
        {...register(props.name, {
          disabled: props.disabled,
          required: props.isRequired,
          // any required to force this initial value to work.
          value: true as any,
          validate: props.validate,
        })}
      />
    );
  } else {
    return (
      <RadioGroup defaultValue={currentDefault}>
        <Stack direction={"row"}>
          <Radio
            {...register(props.name, {
              disabled: props.disabled,
              required: props.isRequired,
              validate: props.validate,
            })}
            value={"true"}
          >
            True
          </Radio>
          <Radio
            {...register(props.name, {
              disabled: props.disabled,
              required: props.isRequired,
              validate: props.validate,
            })}
            value={"false"}
          >
            False
          </Radio>
        </Stack>
      </RadioGroup>
    );
  }
};
