import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  ButtonGroup,
  Editable,
  EditableInput,
  EditablePreview,
  Flex,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  InputGroup,
  InputRightElement,
  Progress,
  Switch,
  Text,
  Tooltip,
} from "@chakra-ui/react";
import { debounce, throttle } from "lodash";
import { ChangeEvent, MutableRefObject, ReactElement, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Virtuoso, VirtuosoHandle } from "react-virtuoso";
import { isDefined, isNotEmpty, stringifyError, stripAnsi } from "../../../utils";
import { CopyButton } from "../../CopyButton";
import { DownloadButton } from "../../DownloadButton";
import { LogLine } from "./LogLine";
import { LogLineMessage } from "./types";
import { normalizeLogText } from "./utils";

type LogViewerProps = {
  logLines: LogLineMessage[];
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

  const searchRef: MutableRefObject<HTMLInputElement | null> = useRef(null);
  const [rawSearchTerm, setRawSearchTerm] = useState("");
  const [maybeSearchPattern, setMaybeSearchPattern] = useState<{ pattern?: RegExp; error?: string }>({});
  const [searchMatchesIndices, setSearchMatchesIndices] = useState<number[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState<number | undefined>(undefined);

  const throttledSetLogLines = useMemo(() => throttle(setLogLines, 500), []);

  const updateMatches = useCallback(
    (searchTerm: string) => {
      if (isNotEmpty(searchTerm)) {
        try {
          const pattern = new RegExp(searchTerm, "gi"); // i is case insensitive
          const matches = logLines
            .map((line, index) => {
              if (line?.message && normalizeLogText(line.message).match(pattern)) {
                return index;
              }
              return null;
            })
            .filter(isDefined);
          setMaybeSearchPattern({ pattern });
          setSearchMatchesIndices(matches);
          setCurrentSearchIndex(matches.length > 0 ? 0 : undefined);
        } catch (error: any) {
          setMaybeSearchPattern({ error: stringifyError(error) });
          setSearchMatchesIndices([]);
          setCurrentSearchIndex(undefined);
        }
      } else {
        setSearchMatchesIndices([]);
        setCurrentSearchIndex(undefined);
      }
    },
    [logLines],
  );

  const debouncedUpdateMatches = useMemo(() => debounce(updateMatches, 100), [updateMatches]);

  const handleOnChange = (e: ChangeEvent<HTMLInputElement>) => {
    setRawSearchTerm(e.target.value);
    debouncedUpdateMatches(e.target.value);
  };

  const updateSearchIndexBounded = (newIndex: number) => {
    if (newIndex > searchMatchesIndices.length - 1) {
      newIndex = 0;
    }
    if (newIndex < 0) {
      newIndex = searchMatchesIndices.length - 1;
    }
    setCurrentSearchIndex(newIndex);
    virtuosoRef.current?.scrollToIndex(searchMatchesIndices[newIndex]);
  };

  const handlePriorMatchClick = () => {
    updateSearchIndexBounded(isDefined(currentSearchIndex) ? currentSearchIndex - 1 : 0);
  };

  const handleNextMatchClick = () => {
    updateSearchIndexBounded(isDefined(currentSearchIndex) ? currentSearchIndex + 1 : 0);
  };

  const handleClearSearch = () => {
    setRawSearchTerm("");
    setMaybeSearchPattern({});
    setSearchMatchesIndices([]);
    setCurrentSearchIndex(undefined);
  };

  const handleIndexInputChange = (text: string) => {
    let index = parseInt(text);
    if (isNaN(index)) {
      index = 1;
    }
    if (index > searchMatchesIndices.length) {
      index = searchMatchesIndices.length;
    }
    updateSearchIndexBounded(index - 1);
  };

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

  const isIndexSelected = (index: number) => {
    return isDefined(currentSearchIndex) && searchMatchesIndices[currentSearchIndex] === index;
  };

  useEffect(() => {
    const listener = function (e: KeyboardEvent) {
      const element = searchRef?.current;
      if ((e.ctrlKey && e.keyCode === 70) || (e.metaKey && e.keyCode === 70)) {
        if (element !== document.activeElement) {
          e.preventDefault();
          element?.focus();
        }
      }
      // Next search match with cmd/ctrl+G
      // if ((e.ctrlKey && e.keyCode === 71) || (e.metaKey && e.keyCode === 71)) {
      //   console.log("NEXT", e.keyCode);
      //   e.preventDefault();
      //   nextMatch();
      // }

      // Clear the search on escape
      if (e.key === "Escape" || e.keyCode === 27) {
        if (element === document.activeElement) {
          e.preventDefault();
          setRawSearchTerm("");
          setSearchMatchesIndices([]);
          setCurrentSearchIndex(undefined);
        }
      }
    };
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, []);

  useEffect(() => {
    throttledSetLogLines(propsLogLines);
  }, [propsLogLines, throttledSetLogLines]);

  return (
    <Flex flexDirection={"column"} gap={"32px"} h={"100%"}>
      <Flex flexDirection={"column"} position={"relative"} bg={"gray.800"} h={"100%"}>
        <Box width={"100%"}>
          <Flex m={4}>
            <FormControl isInvalid={isDefined(maybeSearchPattern.error)}>
              <Flex>
                <InputGroup size="sm" width={"40%"}>
                  <Input
                    size={"sm"}
                    ref={searchRef}
                    value={rawSearchTerm}
                    onChange={handleOnChange}
                    placeholder={"search"}
                  />
                  {rawSearchTerm && (
                    <InputRightElement>
                      <SmallCloseIcon onClick={handleClearSearch} />
                    </InputRightElement>
                  )}
                </InputGroup>
                <Button
                  size={"sm"}
                  ml={2}
                  onClick={handlePriorMatchClick}
                  isDisabled={searchMatchesIndices.length === 0}
                  colorScheme={"darkBlue"}
                >
                  Previous
                </Button>
                <Button
                  size={"sm"}
                  ml={2}
                  onClick={handleNextMatchClick}
                  isDisabled={searchMatchesIndices.length === 0}
                  colorScheme={"darkBlue"}
                >
                  Next
                </Button>
                {rawSearchTerm.length > 0 && (
                  <Flex ml={2} alignItems={"center"}>
                    <Text align={"left"} color={searchMatchesIndices.length === 0 ? "red" : "kurtosisGreen.400"}>
                      {searchMatchesIndices.length > 0 && currentSearchIndex !== undefined && (
                        <span>
                          <Editable
                            display={"inline"}
                            p={0}
                            m={"0 4px 0 0"}
                            size={"sm"}
                            value={`${currentSearchIndex + 1}`}
                            onChange={handleIndexInputChange}
                          >
                            <Tooltip label="Click to edit" shouldWrapChildren={true}>
                              <EditablePreview />
                            </Tooltip>
                            <EditableInput p={1} width={"50px"} />
                          </Editable>
                          <>/ </>
                        </span>
                      )}
                      <span>{searchMatchesIndices.length} matches</span>
                    </Text>
                  </Flex>
                )}
              </Flex>
              {isDefined(maybeSearchPattern.error) && <FormErrorMessage>{maybeSearchPattern.error}</FormErrorMessage>}
            </FormControl>
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
          itemContent={(index, line) => (
            <LogLine {...line} highlightPattern={maybeSearchPattern.pattern} selected={isIndexSelected(index)} />
          )}
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
