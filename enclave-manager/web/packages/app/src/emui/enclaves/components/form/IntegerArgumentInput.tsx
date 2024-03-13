import { Input } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";

import { useFormContext } from "react-hook-form";
import { KurtosisFormInputProps } from "./types";

export const IntegerArgumentInput = <DataModel extends object>(props: KurtosisFormInputProps<DataModel>) => {
  const { register } = useFormContext<DataModel>();

  return (
    <Input
      {...register(props.name, {
        required: props.isRequired,
        validate: (value: number) => {
          if (isNaN(value)) {
            return "This value should be an integer";
          }

          const propsValidation = props.validate ? props.validate(value) : undefined;
          if (isDefined(propsValidation)) {
            return propsValidation;
          }
        },
      })}
      isReadOnly={props.disabled}
      placeholder={props.placeholder}
      width={props.width}
      size={props.size || "lg"}
      tabIndex={props.tabIndex}
    />
  );
};
