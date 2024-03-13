import {
  Badge,
  Flex,
  FormControl,
  FormControlProps,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  Tag,
} from "@chakra-ui/react";
import { isDefined, KurtosisMarkdown } from "kurtosis-ui-components";
import { PropsWithChildren } from "react";
import { FieldError, FieldPath, useFormContext } from "react-hook-form";

type KurtosisFormControlProps<DataModel extends object> = PropsWithChildren<
  FormControlProps & {
    name: FieldPath<DataModel>;
    label: string;
    type?: string;
    helperText?: string;
    disabled?: boolean;
    isRequired?: boolean;
  }
>;
export const KurtosisFormControl = <DataModel extends object>({
  name,
  label,
  type,
  helperText,
  disabled,
  isRequired,
  children,
  ...formControlProps
}: KurtosisFormControlProps<DataModel>) => {
  const {
    formState: { errors },
  } = useFormContext<DataModel>();
  // This looks a little strange because `FieldErrors` has the same structure as `ConfigureEnclaveForm`
  const error = name
    .split(".")
    .reduce((e, part) => (isDefined(e) ? e[part] : undefined), errors as Record<string, any>) as FieldError | undefined;

  return (
    <FormControl isInvalid={isDefined(error)} isDisabled={disabled} isRequired={isRequired} {...formControlProps}>
      <Flex justifyContent={"space-between"}>
        <Flex alignItems={"center"}>
          <FormLabel fontWeight={"bold"}>{label}</FormLabel>
          {isDefined(type) && <Badge mb={2}>{type}</Badge>}
        </Flex>
        <Flex flexDirection={"column"} justifyContent={"center"} alignItems={"center"}>
          <Tag colorScheme={isRequired ? "red" : "gray"} variant={"square"}>
            {isRequired ? "Required" : "Optional"}
          </Tag>
        </Flex>
      </Flex>
      {children}
      <FormHelperText>
        <KurtosisMarkdown>{helperText}</KurtosisMarkdown>
      </FormHelperText>
      <FormErrorMessage>{error?.type === "required" ? "This field is required" : error?.message}</FormErrorMessage>
    </FormControl>
  );
};

type KurtosisSubtypeFormControlProps<DataModel extends object> = PropsWithChildren<{
  name: FieldPath<DataModel>;
  disabled?: boolean;
  isRequired?: boolean;
}> &
  FormControlProps;
export const KurtosisSubtypeFormControl = <DataModel extends object>({
  name,
  disabled,
  isRequired,
  children,
  ...formControlProps
}: KurtosisSubtypeFormControlProps<DataModel>) => {
  const {
    formState: { errors },
  } = useFormContext<DataModel>();
  // This looks a little strange because `FieldErrors` has the same structure as `ConfigureEnclaveForm`
  const error = name
    .split(".")
    .reduce((e, part) => (isDefined(e) ? e[part] : undefined), errors as Record<string, any>) as FieldError | undefined;

  return (
    <FormControl
      {...formControlProps}
      width={"unset"}
      isInvalid={isDefined(error)}
      isDisabled={disabled}
      isRequired={isRequired}
    >
      {children}
    </FormControl>
  );
};
