import { Badge, Flex, FormControl, FormErrorMessage, FormHelperText, FormLabel } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { FieldError, FieldPath } from "react-hook-form";
import Markdown from "react-markdown";
import { isDefined } from "../../../utils";
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
        <Markdown>{helperText}</Markdown>
      </FormHelperText>
      <FormErrorMessage>{error?.message}</FormErrorMessage>
    </FormControl>
  );
};
