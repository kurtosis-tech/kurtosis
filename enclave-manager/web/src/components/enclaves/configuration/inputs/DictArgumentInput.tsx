import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";
import { ArgumentValueType } from "../../../../client/packageIndexer/api/kurtosis_package_indexer_pb";

import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { stringifyError } from "../../../../utils";
import { CopyButton } from "../../../CopyButton";
import { PasteButton } from "../../../PasteButton";
import { ConfigureEnclaveForm } from "../types";
import { KurtosisArgumentTypeInput, KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

type DictArgumentInputProps = Omit<KurtosisArgumentTypeInputProps, "type"> & {
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
          <KurtosisArgumentTypeInput
            type={keyType}
            name={`${otherProps.name as `args.${string}.${number}.value`}.${i}.key`}
            validate={otherProps.validate}
            isRequired
            size={"xs"}
            width={"222px"}
          />
          <KurtosisArgumentTypeInput
            type={valueType}
            name={`${otherProps.name as `args.${string}.${number}.value`}.${i}.value`}
            validate={otherProps.validate}
            isRequired
            size={"xs"}
            width={"222px"}
          />
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"xs"} colorScheme={"red"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button onClick={() => append({})} leftIcon={<FiPlus />} size={"xs"} colorScheme={"kurtosisGreen"}>
          Add
        </Button>
      </Flex>
    </Flex>
  );
};
