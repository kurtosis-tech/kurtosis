import { Box, ButtonGroup, Flex, FormControl, FormLabel, Progress, Switch } from "@chakra-ui/react";
import { throttle } from "lodash";
import { ChangeEvent, ReactElement, useEffect, useMemo, useRef, useState } from "react";
import { Virtuoso, VirtuosoHandle } from "react-virtuoso";
import { isDefined, stripAnsi } from "../../../utils";
import { CopyButton } from "../../CopyButton";
import { DownloadButton } from "../../DownloadButton";
import { LogLine, LogLineProps } from "./LogLine";

type LogViewerProps = {
  logLines: LogLineProps[];
  progressPercent?: number | "indeterminate" | "failed";
  ProgressWidget?: ReactElement;
  logsFileName?: string;
};

export const LogViewer = ({
  progressPercent,
  logLines: propsLogLines,
  ProgressWidget,
  logsFileName,
}: LogViewerProps) => {
  const virtuosoRef = useRef<VirtuosoHandle>(null);
  const [logLines, setLogLines] = useState(propsLogLines);
  const [userIsScrolling, setUserIsScrolling] = useState(false);
  const [automaticScroll, setAutomaticScroll] = useState(true);

  const throttledSetLogLines = useMemo(() => throttle(setLogLines, 500), []);

  useEffect(() => {
    throttledSetLogLines(propsLogLines);
  }, [propsLogLines, throttledSetLogLines]);

  const handleAutomaticScrollChange = (e: ChangeEvent<HTMLInputElement>) => {
    setAutomaticScroll(e.target.checked);
    if (virtuosoRef.current && e.target.checked) {
      virtuosoRef.current.scrollToIndex({ index: "LAST" });
    }
  };

  const handleBottomStateChange = (atBottom: boolean) => {
    if (userIsScrolling) {
      setAutomaticScroll(atBottom);
    } else if (automaticScroll && !atBottom) {
      virtuosoRef.current?.scrollToIndex({ index: "LAST" });
    }
  };

  const getLogsValue = () => {
    return (
      logLines
        .map(({ message }) => message)
        .filter(isDefined)
        .map(stripAnsi)
        .join("\n")
    );
  };

  return (
    <Flex flexDirection={"column"} gap={"32px"}>
      <Flex flexDirection={"column"} position={"relative"} bg={"gray.800"}>
        {isDefined(ProgressWidget) && (
          <Box
            display={"inline-flex"}
            alignItems={"center"}
            justifyContent={"center"}
            gap={"8px"}
            position={"absolute"}
            top={"16px"}
            right={"16px"}
            padding={"24px"}
            h={"48px"}
            bg={"gray.650"}
            borderRadius={"8px"}
            fontSize={"xl"}
            fontWeight={"semibold"}
            zIndex={1}
          >
            {ProgressWidget}
          </Box>
        )}
        <Virtuoso
          ref={virtuosoRef}
          followOutput={automaticScroll}
          atBottomStateChange={handleBottomStateChange}
          isScrolling={setUserIsScrolling}
          style={{ height: "660px" }}
          data={logLines.filter(({ message }) => isDefined(message))}
          itemContent={(_, line) => <LogLine {...line} />}
        />
        {isDefined(progressPercent) && (
          <Progress
            value={typeof progressPercent === "number" ? progressPercent : progressPercent === "failed" ? 100 : 0}
            isIndeterminate={progressPercent === "indeterminate"}
            height={"4px"}
            colorScheme={progressPercent === "failed" ? "red.500" : "kurtosisGreen"}
          />
        )}
      </Flex>
      <Flex alignItems={"space-between"} width={"100%"}>
        <FormControl display={"flex"} alignItems={"center"}>
          <Switch isChecked={automaticScroll} onChange={handleAutomaticScrollChange} />
          <FormLabel mb={"0"} marginInlineStart={3}>
            Automatic Scroll
          </FormLabel>
        </FormControl>
        <ButtonGroup isAttached>
          <CopyButton contentName={"logs"} valueToCopy={getLogsValue} size={"md"} isDisabled={logLines.length === 0} />
          <DownloadButton
            valueToDownload={getLogsValue}
            size={"md"}
            fileName={logsFileName || `logs.txt`}
            isDisabled={logLines.length === 0}
          />
        </ButtonGroup>
      </Flex>
    </Flex>
  );
};
