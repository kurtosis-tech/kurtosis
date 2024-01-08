import { Box, Flex } from "@chakra-ui/react";

import Convert from "ansi-to-html";
import parse from "html-react-parser";
import { ReactElement } from "react";
import { hasAnsi, isDefined } from "../utils";
import { LogLineMessage, LogStatus } from "./types";
import { normalizeLogText } from "./utils";

const convert = new Convert();

type LogLineProps = LogLineMessage & {
  highlightPattern?: RegExp;
  selected?: boolean;
};

const logFontFamily = "Menlo, Monaco, Inconsolata, Consolas, Courier, monospace";

export const LogLine = ({ timestamp, message, status, highlightPattern, selected }: LogLineProps) => {
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
    <Flex p={"2px 0"} m={"0 16px"} gap={"8px"} alignItems={"top"} backgroundColor={selected ? "gray.600" : ""}>
      {isDefined(timestamp) && (
        <Box
          as={"pre"}
          whiteSpace={"pre-wrap"}
          fontSize={"xs"}
          lineHeight="2"
          fontWeight={600}
          fontFamily={logFontFamily}
          color={"grey"}
          minW={"140px"}
        >
          <>{timestamp.toLocal().toFormat("yyyy-MM-dd HH:mm:ss ZZZZ")}</>
        </Box>
      )}
      <Box
        as={"pre"}
        whiteSpace={"pre-wrap"}
        overflowWrap={"anywhere"}
        fontSize={"xs"}
        lineHeight="2"
        fontWeight={400}
        fontFamily={logFontFamily}
        color={statusToColor(status)}
        _focus={{ boxShadow: "outline" }}
      >
        <Message message={message} highlightPattern={highlightPattern} />
      </Box>
    </Flex>
  );
};

type MessageProps = {
  message?: string;
  highlightPattern?: RegExp;
};

const Message = ({ message, highlightPattern }: MessageProps) => {
  if (!isDefined(message)) {
    return null;
  }

  if (hasAnsi(message)) {
    return <>{parse(convert.toHtml(message))}</>;
  }

  if (highlightPattern) {
    const normalizedLogText = normalizeLogText(message);
    const splitText = normalizedLogText.split(highlightPattern);
    const matches = normalizedLogText.match(highlightPattern);

    if (!isDefined(matches)) {
      return <span>{message}</span>;
    }

    return (
      <span>
        {splitText.reduce(
          (arr: (ReactElement | string)[], element, index) =>
            matches[index] ? [...arr, element, <mark key={index}>{matches[index]}</mark>] : [...arr, element],
          [],
        )}
      </span>
    );
  }

  return <>{message}</>;
};
