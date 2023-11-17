import {
  Box,
  Button,
  ButtonGroup,
  Flex,
  FormControl,
  FormLabel,
  Input,
  InputGroup,
  Progress,
  Switch,
  Text,
} from "@chakra-ui/react";
import { debounce, throttle } from "lodash";
import { ChangeEvent, ReactElement, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Virtuoso, VirtuosoHandle } from "react-virtuoso";
import { isDefined, isNotEmpty, stripAnsi } from "../../../utils";
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
  const [searchTerm, setSearchTerm] = useState("");
  const [totalSearchMatches, setTotalSearchMatches] = useState<number | undefined>(undefined);

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
    return logLines
      .map(({ message }) => message)
      .filter(isDefined)
      .map(stripAnsi)
      .join("\n");
  };

  const updateSearchTerm = (text: string) => {
    console.log(`updating search term: ${text}`);
    setSearchTerm(text);
  };
  const debouncedUpdateSearchTerm = debounce(updateSearchTerm, 500);
  const debouncedUpdateSearchTermCallback = useCallback(debouncedUpdateSearchTerm, []);

  const findMatches = (searchTerm: string) => {
    if (hasSearchTerm()) {
      setTotalSearchMatches(undefined);
      let counter = 0;
      console.log(`looking for "${searchTerm}"`);
      logLines.forEach((line, index) => {
        if (line.message?.match(searchTerm)) {
          counter++;
          // console.log(`Found match for ${searchTerm} on line ${index + 1}`);
        }
      });
      console.log(`Found ${counter} matches for '${searchTerm}'`);
      setTotalSearchMatches(counter);
    }
  };

  useEffect(() => {
    findMatches(searchTerm);
  }, [searchTerm]);

  const hasSearchTerm = () => {
    return isDefined(searchTerm) && isNotEmpty(searchTerm);
  };

  return (
    <Flex flexDirection={"column"} gap={"32px"} h={"100%"}>
      <Flex flexDirection={"column"} position={"relative"} bg={"gray.800"} h={"100%"}>
        <Box width={"100%"}>
          <Flex m={4}>
            <Flex width={"40%"}>
              <InputGroup>
                <Input placeholder={"search"} onChange={(e) => debouncedUpdateSearchTermCallback(e.target.value)} />
                {/*<InputRightElement>*/}
                {/*  <BsRegex />*/}
                {/*</InputRightElement>*/}
              </InputGroup>
            </Flex>
            <Button ml={2}>Previous</Button>
            <Button ml={2}>Next</Button>
            {hasSearchTerm() && isDefined(totalSearchMatches) && (
              <Box ml={2}>
                <Text align={"center"}>
                  {totalSearchMatches} matches for "{searchTerm}"
                </Text>
              </Box>
            )}
          </Flex>
        </Box>
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
          style={{ height: "100%" }}
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
