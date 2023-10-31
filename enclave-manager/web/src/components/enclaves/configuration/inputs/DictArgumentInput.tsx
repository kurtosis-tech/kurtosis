import { Button, Flex } from "@chakra-ui/react";
import { ArgumentValueType } from "../../../../client/packageIndexer/api/kurtosis_package_indexer_pb";

import { useFieldArray } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisArgumentTypeInput, KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

type DictArgumentInputProps = Omit<KurtosisArgumentTypeInputProps, "type"> & {
  keyType: ArgumentValueType;
  valueType: ArgumentValueType;
};

export const DictArgumentInput = ({ keyType, valueType, ...otherProps }: DictArgumentInputProps) => {
  const { fields, append, remove } = useFieldArray({ name: otherProps.name });

  return (
    <Flex flexDirection={"column"} gap={"10px"}>
      {fields.map((field, i) => (
        <Flex key={i} gap={"10px"}>
          <KurtosisArgumentTypeInput
            type={keyType}
            name={`${otherProps.name as `args.${string}.${number}.value`}.${i}.key`}
            isRequired
          />
          <KurtosisArgumentTypeInput
            type={valueType}
            name={`${otherProps.name as `args.${string}.${number}.value`}.${i}.value`}
            isRequired
          />
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} minW={"90px"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button onClick={() => append({})} leftIcon={<FiPlus />} minW={"90px"}>
          Add
        </Button>
      </Flex>
    </Flex>
  );
};
