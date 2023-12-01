import { Badge, Flex, FormControl, FormErrorMessage, FormHelperText, FormLabel } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { FieldError, FieldPath } from "react-hook-form";
import { isDefined } from "../../../utils";
import { KurtosisMarkdown } from "../../KurtosisMarkdown";
import { useEnclaveConfigurationFormContext } from "./EnclaveConfigurationForm";
import { ConfigureEnclaveForm } from "./types";

type KurtosisArguementFormControlProps = PropsWithChildren<{
  name: FieldPath<ConfigureEnclaveForm>;
  label: string;
  type: string;
  helperText?: string;
  disabled?: boolean;
  isRequired?: boolean;
}>;
export const KurtosisArgumentFormControl = ({
  name,
  label,
  type,
  helperText,
  disabled,
  isRequired,
  children,
}: KurtosisArguementFormControlProps) => {
  const {
    formState: { errors },
  } = useEnclaveConfigurationFormContext();
  // This looks a little strange because `FieldErrors` has the same structure as `ConfigureEnclaveForm`
  const error = name
    .split(".")
    .reduce((e, part) => (isDefined(e) ? e[part] : undefined), errors as Record<string, any>) as FieldError | undefined;

  return (
    <FormControl isInvalid={isDefined(error)} isDisabled={disabled} isRequired={isRequired}>
      <Flex alignItems={"center"}>
        <FormLabel>{label}</FormLabel>
        <Badge mb={2}>{type}</Badge>
      </Flex>
      {children}
      <FormHelperText>
        <KurtosisMarkdown>{helperText}</KurtosisMarkdown>
      </FormHelperText>
      <FormErrorMessage>{error?.message}</FormErrorMessage>
    </FormControl>
  );
};

type KurtosisArguementSubtypeFormControlProps = PropsWithChildren<{
  name: FieldPath<ConfigureEnclaveForm>;
  disabled?: boolean;
  isRequired?: boolean;
}>;
export const KurtosisArgumentSubtypeFormControl = ({
  name,
  disabled,
  isRequired,
  children,
}: KurtosisArguementSubtypeFormControlProps) => {
  const {
    formState: { errors },
  } = useEnclaveConfigurationFormContext();
  // This looks a little strange because `FieldErrors` has the same structure as `ConfigureEnclaveForm`
  const error = name
    .split(".")
    .reduce((e, part) => (isDefined(e) ? e[part] : undefined), errors as Record<string, any>) as FieldError | undefined;

  return (
    <FormControl width={"unset"} isInvalid={isDefined(error)} isDisabled={disabled} isRequired={isRequired}>
      {children}
    </FormControl>
  );
};
