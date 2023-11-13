import { Radio, RadioGroup, Stack, Switch } from "@chakra-ui/react";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

type BooleanArgumentInputProps = Omit<KurtosisArgumentTypeInputProps, "type"> & {
  inputType?: "radio" | "switch";
};

export const BooleanArgumentInput = ({ inputType, ...props }: BooleanArgumentInputProps) => {
  const { register, getValues } = useEnclaveConfigurationFormContext();

  const currentDefault = getValues(props.name);

  if (inputType === "switch") {
    return (
      <Switch
        {...register(props.name, {
          disabled: props.disabled,
          required: props.isRequired,
          value: true,
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
