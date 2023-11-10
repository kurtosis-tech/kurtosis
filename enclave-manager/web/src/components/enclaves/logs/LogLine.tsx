import { Box, Flex } from "@chakra-ui/react";
import { DateTime } from "luxon";
import { isDefined } from "../../../utils";
import parse from "html-react-parser";
// @ts-ignore
import hasAnsi from "has-ansi";


const Convert = require('ansi-to-html');
const convert = new Convert();

export type LogStatus = "info" | "error";

export type LogLineProps = {
  timestamp?: DateTime;
  message?: string;
  status?: LogStatus;
};

const logFontFamily = "Menlo, Monaco, Inconsolata, Consolas, Courier, monospace";

export const LogLine = ({ timestamp, message, status }: LogLineProps) => {
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

  const processText = (message: string) => {
    if (hasAnsi(message)) {
      const ansiInHtml = convert.toHtml(message)
      return parse(ansiInHtml);
    } else {
      return <>{message}</>
    }
  }

  return (
    <Flex p={"2px 0"} m={"0 16px"} gap={"8px"} alignItems={"top"}>
      {isDefined(timestamp) && (
        <Box
          as={"pre"}
          whiteSpace={"pre-wrap"}
          fontSize={"xs"}
          lineHeight="2"
          fontWeight={600}
          fontFamily={logFontFamily}
          color={"grey"}
          minW={"200px"}
        >
          <>{timestamp.toLocal().toFormat("yyyy-MM-dd HH:mm:ss.SSS ZZZZ")}</>
        </Box>
      )}
      <Box
        as={"pre"}
        whiteSpace={"pre-wrap"}
        fontSize={"xs"}
        lineHeight="2"
        fontWeight={400}
        fontFamily={logFontFamily}
        color={statusToColor(status)}
      >
        {message && processText(message)}
      </Box>
    </Flex>
  );
};
