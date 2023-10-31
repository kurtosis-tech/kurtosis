import { Alert, AlertDescription, AlertIcon, AlertTitle } from "@chakra-ui/react";

type KurtosisAlertProps = {
  message: string;
};

export const KurtosisAlert = ({ message }: KurtosisAlertProps) => {
  return (
    <Alert status="error">
      <AlertIcon />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
};
