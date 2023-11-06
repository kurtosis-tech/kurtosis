import { Box } from "@chakra-ui/react";

export type LogStatus = "info" | "error";

export type LogLineProps = {
  message?: string;
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
        return "white";
    }
  };

  return (
    <Box
      as={"pre"}
      whiteSpace={"pre-wrap"}
      borderBottom={"1px solid #444444"}
      p={"14px 0"}
      m={"0 16px"}
      fontSize={"xs"}
      lineHeight="2"
      fontWeight={400}
      fontFamily="Ubuntu Mono"
      color={statusToColor(status)}
    >
      {message || <i>No message</i>}
    </Box>
  );
};
