import { Button, ButtonGroup, Flex, useToast } from "@chakra-ui/react";

import { CopyButton, PasteButton, stringifyError } from "kurtosis-ui-components";
import { FC } from "react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { FiDelete, FiPlus } from "react-icons/fi";
import { KurtosisSubtypeFormControl } from "./KurtosisFormControl";
import { KurtosisFormInputProps } from "./types";

type DictArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  KeyFieldComponent: FC<KurtosisFormInputProps<DataModel>>;
  ValueFieldComponent: FC<KurtosisFormInputProps<DataModel>>;
};

export const DictArgumentInput = <DataModel extends object>({
  KeyFieldComponent,
  ValueFieldComponent,
  ...otherProps
}: DictArgumentInputProps<DataModel>) => {
  const toast = useToast();
  const { getValues, setValue } = useFormContext<DataModel>();
  const { fields, append, remove } = useFieldArray({ name: otherProps.name });

  const handleValuePaste = (value: string) => {
    try {
      const parsed = JSON.parse(value);
      setValue(otherProps.name, Object.entries(parsed).map(([key, value]) => ({ key, value })) as any);
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
              (getValues(otherProps.name) as any[]).reduce(
                (acc: Record<string, any>, { key, value }: { key: string; value: any }) => ({ ...acc, [key]: value }),
                {},
              ),
            )
          }
        />
        <PasteButton onValuePasted={handleValuePaste} />
      </ButtonGroup>
      {fields.map((field, i) => (
        <Flex key={field.id} gap={"10px"}>
          <KurtosisSubtypeFormControl
            name={`${otherProps.name as `args.${string}`}.${i}.key`}
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
          >
            <KeyFieldComponent
              name={`${otherProps.name}.${i}.key` as any}
              validate={otherProps.validate}
              isRequired={true}
              size={"sm"}
              width={"222px"}
            />
          </KurtosisSubtypeFormControl>
          <KurtosisSubtypeFormControl
            name={`${otherProps.name as `args.${string}`}.${i}.value`}
            disabled={otherProps.disabled}
            isRequired={otherProps.isRequired}
          >
            <ValueFieldComponent
              name={`${otherProps.name}.${i}.value` as any}
              validate={otherProps.validate}
              isRequired={true}
              size={"sm"}
              width={"222px"}
            />
          </KurtosisSubtypeFormControl>
          <Button onClick={() => remove(i)} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
            Delete
          </Button>
        </Flex>
      ))}
      <Flex>
        <Button
          onClick={() => append({} as any)}
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
