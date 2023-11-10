import { Alert, AlertDescription, AlertIcon, AlertProps, AlertTitle } from "@chakra-ui/react";

type KurtosisAlertProps = AlertProps & {
  message: string;
};

export const KurtosisAlert = ({ message, ...alertProps }: KurtosisAlertProps) => {
  return (
    <Alert status="error" {...alertProps}>
      <AlertIcon />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
};
