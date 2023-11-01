import { Box } from "@chakra-ui/react";

export type LogStatus = "info" | "error";

export type LogLineProps = {
  message: string;
  status?: LogStatus;
};

export const LogLine = ({ message, status }: LogLineProps) => {
  const statusToColor = (status?: LogStatus) => {
    switch (status) {
      case "error":
        return "red.400";
      case "info":
        return "gray.100";
      default:
        return "gray.100";
    }
  };

  return <Box color={statusToColor(status)}>{message || <i>No message</i>}</Box>;
};
