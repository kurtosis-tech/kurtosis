import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";

import { CopyButton, PasteButton, stringifyError } from "kurtosis-ui-components";
import { ReactElement } from "react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisSubtypeFormControl } from "./KurtosisFormControl";
import { KurtosisFormInputProps } from "./types";

type ListArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  renderFieldInput: (props: KurtosisFormInputProps<DataModel>) => ReactElement;
};

export const ListArgumentInput = <DataModel extends object>({
  renderFieldInput,
  ...otherProps
}: ListArgumentInputProps<DataModel>) => {
  const toast = useToast();
  const { getValues, setValue } = useFormContext<DataModel>();
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
          valueToCopy={() =>
            JSON.stringify((getValues(otherProps.name) as any[]).map(({ value }: { value: any }) => value))
          }
        />
        <PasteButton onValuePasted={handleValuePaste} />
      </ButtonGroup>
      {fields.map((field, i) => (
        <Flex key={field.id} gap={"10px"}>
          <KurtosisSubtypeFormControl
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
            name={`${otherProps.name as `args.${string}`}.${i}.value`}
          >
            {renderFieldInput({
              name: `${otherProps.name}.${i}.value` as any,
              isRequired: true,
              validate: otherProps.validate,
              width: "411px",
              size: "sm",
            })}
          </KurtosisSubtypeFormControl>
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button
          onClick={() => append({ value: "" } as any)}
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
