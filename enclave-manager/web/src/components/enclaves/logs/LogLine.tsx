import { Box, Flex } from "@chakra-ui/react";
import parse from "html-react-parser";
import { DateTime } from "luxon";
import { isDefined } from "../../../utils";
// @ts-ignore
import hasAnsi from "has-ansi";
import { ReactElement } from "react";
import { normalizeLogText } from "./LogViewer";

const Convert = require("ansi-to-html");
const convert = new Convert();

export type LogStatus = "info" | "error";

export type LogLineProps = {
  timestamp?: DateTime;
  message?: string;
  status?: LogStatus;
};

export type LogLineSearch = {
  searchTerm: string;
  pattern: RegExp;
};

export type LogLineInput = {
  logLineProps: LogLineProps;
  logLineSearch?: LogLineSearch;
  selected: boolean | undefined;
};

const logFontFamily = "Menlo, Monaco, Inconsolata, Consolas, Courier, monospace";

export const LogLine = ({ logLineProps, logLineSearch, selected }: LogLineInput) => {
  const { timestamp, message, status } = logLineProps;
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

  const processText = (text: string, selected: boolean | undefined) => {
    let reactComponent;
    if (hasAnsi(text)) {
      reactComponent = parse(convert.toHtml(text));
    } else {
      reactComponent = <>{text}</>;
    }

    if (logLineSearch) {
      reactComponent = HighlightPattern({ text, regex: logLineSearch.pattern, selected });
    }
    return reactComponent;
  };

  const HighlightPattern = ({
    text,
    regex,
    selected,
  }: {
    text: string;
    regex: RegExp;
    selected: boolean | undefined;
  }) => {
    const normalizedLogText = normalizeLogText(text);
    const splitText = normalizedLogText.split(regex);
    const matches = normalizedLogText.match(regex);

    if (!isDefined(matches)) {
      return <span>{text}</span>;
    }

    return (
      <span>
        {splitText.reduce(
          (arr: (ReactElement | string)[], element, index) =>
            matches[index]
              ? [
                  ...arr,
                  element,
                  <mark key={index}>
                    {matches[index]}
                  </mark>,
                ]
              : [...arr, element],
          [],
        )}
      </span>
    );
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
          minW={"200px"}
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
        {message && processText(message, selected)}
      </Box>
    </Flex>
  );
};
