import { Radio, RadioGroup, Stack } from "@chakra-ui/react";
import { useEnclaveConfigurationFormContext } from "../EnclaveConfigurationForm";
import { KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

export const BooleanArgumentInput = (props: Omit<KurtosisArgumentTypeInputProps, "type">) => {
  const { register } = useEnclaveConfigurationFormContext();

  return (
    <RadioGroup>
      <Stack direction={"row"}>
        <Radio
          {...register(props.name, {
            disabled: props.disabled,
            required: props.isRequired,
          })}
          value={"true"}
        >
          True
        </Radio>
        <Radio
          {...register(props.name, {
            disabled: props.disabled,
            required: props.isRequired,
          })}
          value={"false"}
        >
          False
        </Radio>
      </Stack>
    </RadioGroup>
  );
};
