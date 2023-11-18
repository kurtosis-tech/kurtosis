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
  FormLabel,
  HStack,
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
import { isDefined, isNotEmpty, stripAnsi } from "../../../utils";
import { CopyButton } from "../../CopyButton";
import { DownloadButton } from "../../DownloadButton";
import { LogLine, LogLineProps, LogLineSearch } from "./LogLine";

type LogViewerProps = {
  logLines: LogLineProps[];
  progressPercent?: number | "indeterminate" | "failed";
  ProgressWidget?: ReactElement;
  logsFileName?: string;
};

export const normalizeLogText = (rawText: string) => {
  return rawText.trim();
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

  const searchRef: MutableRefObject<HTMLInputElement | null> = useRef(null);
  const [search, setSearch] = useState<LogLineSearch | undefined>(undefined);
  const [rawSearchTerm, setRawSearchTerm] = useState("");
  const [searchMatchesIndices, setSearchMatchesIndices] = useState<number[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState<number | undefined>(undefined);

  useEffect(() => {
    window.addEventListener("keydown", function (e) {
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
          setSearch(undefined);
          setRawSearchTerm("");
          setSearchMatchesIndices([]);
          setCurrentSearchIndex(undefined);
        }
      }
    });
  }, []);

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

  useEffect(() => {
    if (search) findMatches(search);
  }, [search?.searchTerm, logLines]);

  const updateSearchTerm = (rawText: string) => {
    setCurrentSearchIndex(undefined);
    const searchTerm = normalizeLogText(rawText);
    const logLineSearch: LogLineSearch = {
      searchTerm: searchTerm,
      pattern: new RegExp(searchTerm, "gi"), // `i` is invariant case
    };
    setSearch(logLineSearch);
  };
  const debouncedUpdateSearchTerm = debounce(updateSearchTerm, 100);
  const debouncedUpdateSearchTermCallback = useCallback(debouncedUpdateSearchTerm, []);

  const hasSearchTerm = () => {
    if (!search) return false;
    return isDefined(search.searchTerm) && isNotEmpty(search.searchTerm);
  };

  const findMatches = (search: LogLineSearch) => {
    setSearchMatchesIndices([]);
    if (hasSearchTerm()) {
      const matches = logLines.flatMap((line, index) => {
        if (line?.message && normalizeLogText(line.message).match(search.pattern)) {
          return index;
        } else {
          return [];
        }
      });
      setSearchMatchesIndices(matches);
    }
  };

  const handleOnChange = (e: ChangeEvent<HTMLInputElement>) => {
    setRawSearchTerm(e.target.value);
    debouncedUpdateSearchTermCallback(e.target.value);
  };

  const priorMatch = () => {
    if (searchMatchesIndices.length > 0) {
      const newIndex = isDefined(currentSearchIndex) ? currentSearchIndex - 1 : 0;
      updateSearchIndexBounded(newIndex);
    }
  };

  const nextMatch = () => {
    if (searchMatchesIndices.length > 0) {
      const newIndex = isDefined(currentSearchIndex) ? currentSearchIndex + 1 : 0;
      updateSearchIndexBounded(newIndex);
    }
  };

  const updateSearchIndexBounded = (newIndex: number) => {
    if (newIndex > searchMatchesIndices.length - 1) {
      newIndex = 0;
    }
    if (newIndex < 0) {
      newIndex = searchMatchesIndices.length - 1;
    }
    setCurrentSearchIndex(newIndex);
    return newIndex;
  };

  useEffect(() => {
    if (virtuosoRef?.current && currentSearchIndex !== undefined && currentSearchIndex >= 0) {
      virtuosoRef.current.scrollToIndex(searchMatchesIndices[currentSearchIndex]);
    }
  }, [currentSearchIndex]);

  const clearSearch = () => {
    setRawSearchTerm("");
    setSearch(undefined);
    setSearchMatchesIndices([]);
    setCurrentSearchIndex(undefined);
  };

  const parseMatchIndexRequest = (input: string) => {
    let parsed = parseInt(input);
    if (isNaN(parsed) || parsed < 1) return 1;
    if (parsed > searchMatchesIndices.length) return searchMatchesIndices.length;
    return parsed;
  };

  const highlight = (currentSearchIndex: number | undefined, thisIndex: number, searchableIndices: number[]) => {
    return (
      currentSearchIndex !== undefined &&
      searchableIndices.length > 0 &&
      searchableIndices[currentSearchIndex] === thisIndex
    );
  };

  return (
    <Flex flexDirection={"column"} gap={"32px"} h={"100%"}>
      <Flex flexDirection={"column"} position={"relative"} bg={"gray.800"} h={"100%"}>
        <Box width={"100%"}>
          <Flex m={4}>
            <Flex width={"40%"}>
              <InputGroup size="sm">
                <Input
                  size={"sm"}
                  ref={searchRef}
                  value={rawSearchTerm}
                  onChange={handleOnChange}
                  placeholder={"search"}
                />
                {rawSearchTerm && (
                  <InputRightElement>
                    <SmallCloseIcon onClick={clearSearch} />
                  </InputRightElement>
                )}
              </InputGroup>
            </Flex>
            <Button size={"sm"} ml={2} onClick={priorMatch}>
              Previous
            </Button>
            <Button size={"sm"} ml={2} onClick={nextMatch}>
              Next
            </Button>
            {hasSearchTerm() && (
              <Box ml={2}>
                <Text align={"left"} color={searchMatchesIndices.length === 0 ? "red" : "kurtosisGreen.400"}>
                  <HStack alignItems={"center"}>
                    <>
                      {searchMatchesIndices.length > 0 && currentSearchIndex !== undefined && (
                        <>
                          <Editable
                            p={0}
                            m={0}
                            size={"sm"}
                            value={`${currentSearchIndex + 1}`}
                            onChange={(inputString) =>
                              updateSearchIndexBounded(parseMatchIndexRequest(inputString) - 1)
                            }
                          >
                            <Tooltip label="Click to edit" shouldWrapChildren={true}>
                              <EditablePreview />
                            </Tooltip>
                            <EditableInput p={1} width={"50px"} />
                          </Editable>
                          <>/ </>
                        </>
                      )}
                      <>{searchMatchesIndices.length} matches</>
                    </>
                  </HStack>
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
          itemContent={(index, line) => (
            <LogLine
              logLineProps={line}
              logLineSearch={search}
              selected={highlight(currentSearchIndex, index, searchMatchesIndices)}
            />
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
