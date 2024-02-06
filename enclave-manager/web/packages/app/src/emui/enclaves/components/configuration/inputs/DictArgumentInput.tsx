import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";
import { ArgumentValueType } from "kurtosis-cloud-indexer-sdk";

import { CopyButton, PasteButton, stringifyError } from "kurtosis-ui-components";
import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisArgumentSubtypeFormControl } from "../KurtosisArgumentFormControl";
import { ConfigureEnclaveForm } from "../types";
import { KurtosisArgumentTypeInput, KurtosisArgumentTypeInputImplProps } from "./KurtosisArgumentTypeInput";

type DictArgumentInputProps = KurtosisArgumentTypeInputImplProps & {
  keyType: ArgumentValueType;
  valueType: ArgumentValueType;
};

export const DictArgumentInput = ({ keyType, valueType, ...otherProps }: DictArgumentInputProps) => {
  const toast = useToast();
  const { getValues, setValue } = useFormContext<ConfigureEnclaveForm>();
  const { fields, append, remove } = useFieldArray({ name: otherProps.name });

  const handleValuePaste = (value: string) => {
    try {
      const parsed = JSON.parse(value);
      setValue(
        otherProps.name,
        Object.entries(parsed).map(([key, value]) => ({ key, value })),
      );
    } catch (err: any) {
      toast({
        title: `Could not read pasted input, was it a json object? Got error: ${stringifyError(err)}`,
        colorScheme: "red",
      });
    }
  };

  return (
    <Flex flexDirection={"column"} gap={"10px"}>
      <ButtonGroup isAttached>
        <CopyButton
          contentName={"value"}
          valueToCopy={() =>
            JSON.stringify(
              getValues(otherProps.name).reduce(
                (acc: Record<string, any>, { key, value }: { key: string; value: any }) => ({ ...acc, [key]: value }),
                {},
              ),
            )
          }
        />
        <PasteButton onValuePasted={handleValuePaste} />
      </ButtonGroup>
      {fields.map((field, i) => (
        <Flex key={i} gap={"10px"}>
          <KurtosisArgumentSubtypeFormControl
            name={`${otherProps.name as `args.${string}`}.${i}.key`}
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
          >
            <KurtosisArgumentTypeInput
              type={keyType}
              name={`${otherProps.name as `args.${string}`}.${i}.key`}
              validate={otherProps.validate}
              isRequired
              size={"sm"}
              width={"222px"}
            />
          </KurtosisArgumentSubtypeFormControl>
          <KurtosisArgumentSubtypeFormControl
            name={`${otherProps.name as `args.${string}`}.${i}.value`}
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
          >
            <KurtosisArgumentTypeInput
              type={valueType}
              name={`${otherProps.name as `args.${string}`}.${i}.value`}
              validate={otherProps.validate}
              isRequired
              size={"sm"}
              width={"222px"}
            />
          </KurtosisArgumentSubtypeFormControl>
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button
          onClick={() => append({})}
          leftIcon={<FiPlus />}
          size={"sm"}
          colorScheme={"kurtosisGreen"}
          variant={"outline"}
        >
          Add
        </Button>
      </Flex>
    </Flex>
  );
};
