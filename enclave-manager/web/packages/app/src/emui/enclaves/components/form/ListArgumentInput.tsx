import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";

import { CopyButton, PasteButton, stringifyError } from "kurtosis-ui-components";
import { ReactElement } from "react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisSubtypeFormControl } from "./KurtosisFormControl";
import { KurtosisFormInputProps } from "./types";

type ListArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  FieldComponent: (props: KurtosisFormInputProps<DataModel>) => ReactElement;
  createNewValue: () => object;
};

export const ListArgumentInput = <DataModel extends object>({
  FieldComponent,
  createNewValue,
  ...otherProps
}: ListArgumentInputProps<DataModel>) => {
  const toast = useToast();
  const { getValues, setValue } = useFormContext<DataModel>();
  const { fields, append, remove } = useFieldArray({ name: otherProps.name });

  const handleValuePaste = (value: string) => {
    try {
      const parsed = JSON.parse(value);
      setValue(otherProps.name, parsed);
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
        <CopyButton contentName={"value"} valueToCopy={() => JSON.stringify(getValues(otherProps.name) as any[])} />
        <PasteButton onValuePasted={handleValuePaste} />
      </ButtonGroup>
      {fields.map((field, i) => (
        <Flex key={field.id} gap={"10px"}>
          <KurtosisSubtypeFormControl
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
            name={`${otherProps.name as `args.${string}`}.${i}`}
          >
            <FieldComponent name={`${otherProps.name}.${i}` as any} isRequired validate={otherProps.validate} />
          </KurtosisSubtypeFormControl>
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button
          onClick={() => append(createNewValue() as any)}
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
