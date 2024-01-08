import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";

import { ArgumentValueType } from "kurtosis-cloud-indexer-sdk";
import { CopyButton, PasteButton, stringifyError } from "kurtosis-ui-components";
import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisArgumentSubtypeFormControl } from "../KurtosisArgumentFormControl";
import { ConfigureEnclaveForm } from "../types";
import { KurtosisArgumentTypeInput, KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

type ListArgumentInputProps = KurtosisArgumentTypeInputImplProps & {
  valueType: ArgumentValueType;
};

export const ListArgumentInput = ({ valueType, ...otherProps }: ListArgumentInputProps) => {
  const toast = useToast();
  const { getValues, setValue } = useFormContext<ConfigureEnclaveForm>();
  const { fields, append, remove } = useFieldArray({ name: otherProps.name });

  const handleValuePaste = (value: string) => {
    try {
      const parsed = JSON.parse(value);
      setValue(
        otherProps.name,
        parsed.map((value: any) => ({ value })),
      );
    } catch (err: any) {
      toast({
        title: `Could not read pasted input, was it a json list of values? Got error: ${stringifyError(err)}`,
        colorScheme: "red",
      });
    }
  };

  return (
    <Flex flexDirection={"column"} gap={"10px"}>
      <ButtonGroup isAttached>
        <CopyButton
          contentName={"value"}
          valueToCopy={() => JSON.stringify(getValues(otherProps.name).map(({ value }: { value: any }) => value))}
        />
        <PasteButton onValuePasted={handleValuePaste} />
      </ButtonGroup>
      {fields.map((field, i) => (
        <Flex key={field.id} gap={"10px"}>
          <KurtosisArgumentSubtypeFormControl
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
            name={`${otherProps.name as `args.${string}`}.${i}.value`}
          >
            <KurtosisArgumentTypeInput
              type={valueType}
              name={`${otherProps.name as `args.${string}`}.${i}.value`}
              isRequired
              validate={otherProps.validate}
              width={"411px"}
              size={"sm"}
            />
          </KurtosisArgumentSubtypeFormControl>
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button
          onClick={() => append({ value: "" })}
          leftIcon={<FiPlus />}
          colorScheme={"kurtosisGreen"}
          size={"sm"}
          variant={"outline"}
        >
          Add
        </Button>
      </Flex>
    </Flex>
  );
};
