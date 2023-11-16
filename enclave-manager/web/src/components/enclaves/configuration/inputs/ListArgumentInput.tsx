import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";
import { ArgumentValueType } from "../../../../client/packageIndexer/api/kurtosis_package_indexer_pb";

import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { stringifyError } from "../../../../utils";
import { CopyButton } from "../../../CopyButton";
import { PasteButton } from "../../../PasteButton";
import { KurtosisArgumentSubtypeFormControl } from "../KurtosisArgumentFormControl";
import { ConfigureEnclaveForm } from "../types";
import { KurtosisArgumentTypeInput, KurtosisArgumentTypeInputProps } from "./KurtosisArgumentTypeInput";

type ListArgumentInputProps = Omit<KurtosisArgumentTypeInputProps, "type"> & {
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
          size={"sm"}
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
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button onClick={() => append({ value: "" })} leftIcon={<FiPlus />} colorScheme={"kurtosisGreen"} size={"sm"}>
          Add
        </Button>
      </Flex>
    </Flex>
  );
};
