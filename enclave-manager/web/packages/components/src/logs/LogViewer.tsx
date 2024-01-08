import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Button,
  ButtonGroup,
  Editable,
  EditableInput,
  EditablePreview,
  Flex,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Icon,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  Progress,
  Switch,
  Text,
  Tooltip,
} from "@chakra-ui/react";
import { throttle } from "lodash";
import { ChangeEvent, MutableRefObject, ReactElement, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { FiSearch } from "react-icons/fi";
import { MdArrowBackIosNew, MdArrowForwardIos } from "react-icons/md";
import { Virtuoso, VirtuosoHandle } from "react-virtuoso";
import { CopyButton } from "../CopyButton";
import { DownloadButton } from "../DownloadButton";
import { FindCommand } from "../KeyboardCommands";
import { useKeyboardAction } from "../useKeyboardAction";
import { isDefined, isNotEmpty, stringifyError, stripAnsi } from "../utils";
import { LogLine } from "./LogLine";
import { LogLineMessage } from "./types";
import { normalizeLogText } from "./utils";

type LogViewerProps = {
  logLines: LogLineMessage[];
  progressPercent?: number | "indeterminate" | "failed";
  ProgressWidget?: ReactElement;
  logsFileName?: string;
  searchEnabled?: boolean;
  copyLogsEnabled?: boolean;
  onGetAllLogs?: () => AsyncIterable<string>;
};

type SearchBaseState = {
  rawSearchTerm: string;
};

type SearchInitState = SearchBaseState & {
  type: "init";
};

type SearchErrorState = SearchBaseState & {
  type: "error";
  error: string;
};

type SearchSuccessState = SearchBaseState & {
  type: "success";
  pattern: RegExp;
  searchMatchesIndices: number[];
  currentSearchIndex?: number;
};

type SearchState = SearchInitState | SearchErrorState | SearchSuccessState;

export const LogViewer = ({
  progressPercent,
  logLines: propsLogLines,
  ProgressWidget,
  logsFileName,
  searchEnabled,
  copyLogsEnabled,
  onGetAllLogs,
}: LogViewerProps) => {
  const virtuosoRef = useRef<VirtuosoHandle>(null);
  const [logLines, setLogLines] = useState(propsLogLines);
  const [userIsScrolling, setUserIsScrolling] = useState(false);
  const [automaticScroll, setAutomaticScroll] = useState(true);

  const [searchState, setSearchState] = useState<SearchState>({ type: "init", rawSearchTerm: "" });

  const throttledSetLogLines = useMemo(() => throttle(setLogLines, 500), []);

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

  const handleSearchStateChange = useCallback((updater: ((prevState: SearchState) => SearchState) | SearchState) => {
    setSearchState((prevState) => {
      const newState = typeof updater === "object" ? updater : updater(prevState);
      if (
        newState.type === "success" &&
        (prevState.type !== "success" || prevState.currentSearchIndex !== newState.currentSearchIndex) &&
        isDefined(newState.currentSearchIndex)
      ) {
        virtuosoRef.current?.scrollToIndex(newState.searchMatchesIndices[newState.currentSearchIndex]);
      }
      return newState;
    });
  }, []);

  const getLogsValue = () => {
    return logLines
      .map(({ message }) => message)
      .filter(isDefined)
      .map(stripAnsi)
      .join("\n");
  };

  const isIndexSelected = (index: number) => {
    return (
      searchState.type === "success" &&
      isDefined(searchState.currentSearchIndex) &&
      searchState.searchMatchesIndices[searchState.currentSearchIndex] === index
    );
  };

  useEffect(() => {
    throttledSetLogLines(propsLogLines);
  }, [propsLogLines, throttledSetLogLines]);

  return (
    <Flex
      flexDirection={"column"}
      h={"100%"}
      w={"100%"}
      flex={"1"}
      borderRadius={"6px"}
      borderColor={"whiteAlpha.300"}
      borderWidth={"1px"}
      borderStyle={"solid"}
      overflow={"clip"}
    >
      <Flex width={"100%"} p={"12px"} bg={"gray.850"} gap={"16px"}>
        {searchEnabled && (
          <SearchControls searchState={searchState} onChangeSearchState={handleSearchStateChange} logLines={logLines} />
        )}
        {isDefined(ProgressWidget) && ProgressWidget}
      </Flex>
      <Flex flexDirection={"column"} position={"relative"} h={"100%"} flex={"1"}>
        <Virtuoso
          ref={virtuosoRef}
          followOutput={automaticScroll}
          atBottomStateChange={handleBottomStateChange}
          isScrolling={setUserIsScrolling}
          style={{ height: "100%", flex: "1" }}
          data={logLines.filter(({ message }) => isDefined(message))}
          itemContent={(index, line) => (
            <LogLine
              {...line}
              highlightPattern={searchState.type === "success" ? searchState.pattern : undefined}
              selected={isIndexSelected(index)}
            />
          )}
        />
        {isDefined(progressPercent) && (
          <Progress
            value={typeof progressPercent === "number" ? progressPercent : progressPercent === "failed" ? 100 : 0}
            isIndeterminate={progressPercent === "indeterminate"}
            height={"4px"}
            colorScheme={progressPercent === "failed" ? "red" : progressPercent === 100 ? "kurtosisGreen" : "blue"}
          />
        )}
      </Flex>
      <Flex alignItems={"space-between"} width={"100%"} p={"12px"} bg={"gray.850"}>
        <FormControl display={"flex"} alignItems={"center"}>
          <Switch isChecked={automaticScroll} onChange={handleAutomaticScrollChange} size={"sm"} />
          <FormLabel mb={"0"} marginInlineStart={3} fontSize={"sm"}>
            Automatic Scroll
          </FormLabel>
        </FormControl>
        <ButtonGroup>
          {copyLogsEnabled && (
            <CopyButton
              contentName={"logs"}
              valueToCopy={getLogsValue}
              size={"sm"}
              isDisabled={logLines.length === 0}
              isIconButton
              aria-label={"Copy logs"}
              color={"gray.100"}
            />
          )}
          <DownloadButton
            valueToDownload={onGetAllLogs || getLogsValue}
            size={"sm"}
            fileName={logsFileName || `logs.txt`}
            isDisabled={logLines.length === 0}
            isIconButton
            aria-label={"Download logs"}
            color={"gray.100"}
          />
        </ButtonGroup>
      </Flex>
    </Flex>
  );
};

type SearchControlsProps = {
  searchState: SearchState;
  onChangeSearchState: (update: ((oldSearchState: SearchState) => SearchState) | SearchState) => void;
  logLines: LogLineMessage[];
};

const SearchControls = ({ searchState, onChangeSearchState, logLines }: SearchControlsProps) => {
  const searchRef: MutableRefObject<HTMLInputElement | null> = useRef(null);
  const [showSearchForm, setShowSearchForm] = useState(false);

  const maybeCurrentSearchIndex = searchState.type === "success" ? searchState.currentSearchIndex : null;

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
          onChangeSearchState((state) => ({
            type: "success",
            rawSearchTerm: state.rawSearchTerm,
            pattern,
            searchMatchesIndices: matches,
            currentSearchIndex: matches.length > 0 ? 0 : undefined,
          }));
        } catch (error: any) {
          onChangeSearchState((state) => ({
            type: "error",
            rawSearchTerm: state.rawSearchTerm,
            error: stringifyError(error),
          }));
        }
      } else {
        onChangeSearchState((state) => ({ type: "init", rawSearchTerm: state.rawSearchTerm }));
      }
    },
    [logLines, onChangeSearchState],
  );

  const throttledUpdateMatches = useMemo(() => throttle(updateMatches, 300), [updateMatches]);

  const handleOnChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChangeSearchState((state) => ({ ...state, rawSearchTerm: e.target.value }));
    throttledUpdateMatches(e.target.value);
  };

  const updateSearchIndexBounded = useCallback(
    (newIndex: number) => {
      onChangeSearchState((searchState) => {
        if (searchState.type !== "success" || searchState.searchMatchesIndices.length === 0) {
          return searchState;
        }
        if (newIndex > searchState.searchMatchesIndices.length - 1) {
          newIndex = 0;
        }
        if (newIndex < 0) {
          newIndex = searchState.searchMatchesIndices.length - 1;
        }
        return { ...searchState, currentSearchIndex: newIndex };
      });
    },
    [onChangeSearchState],
  );

  const handlePriorMatchClick = useCallback(() => {
    updateSearchIndexBounded(isDefined(maybeCurrentSearchIndex) ? maybeCurrentSearchIndex - 1 : 0);
  }, [updateSearchIndexBounded, maybeCurrentSearchIndex]);

  const handleNextMatchClick = useCallback(() => {
    updateSearchIndexBounded(isDefined(maybeCurrentSearchIndex) ? maybeCurrentSearchIndex + 1 : 0);
  }, [updateSearchIndexBounded, maybeCurrentSearchIndex]);

  const handleClearSearch = useCallback(() => {
    onChangeSearchState({ type: "init", rawSearchTerm: "" });
  }, [onChangeSearchState]);

  const handleIndexInputChange = (text: string) => {
    if (searchState.type !== "success") {
      return;
    }
    let index = parseInt(text);
    if (isNaN(index)) {
      index = 1;
    }
    if (index > searchState.searchMatchesIndices.length) {
      index = searchState.searchMatchesIndices.length;
    }
    updateSearchIndexBounded(index - 1);
  };

  useKeyboardAction(
    useMemo(
      () => ({
        find: () => {
          setShowSearchForm(true);
          if (isDefined(searchRef.current) && searchRef.current !== document.activeElement) {
            searchRef.current.focus();
          }
        },
        enter: () => {
          handleNextMatchClick();
        },
        "shift-enter": () => {
          handlePriorMatchClick();
        },
        escape: () => {
          if (isDefined(searchRef.current) && searchRef.current === document.activeElement) {
            handleClearSearch();
          }
        },
      }),
      [searchRef, handlePriorMatchClick, handleNextMatchClick, handleClearSearch],
    ),
  );

  if (!showSearchForm) {
    return (
      <Button
        bg={"gray.650"}
        color={"gray.150"}
        leftIcon={<FiSearch />}
        rightIcon={<FindCommand />}
        variant={"solid"}
        onClick={() => setShowSearchForm(true)}
      >
        Search
      </Button>
    );
  } else {
    return (
      <FormControl isInvalid={searchState.type === "error"}>
        <Flex gap={"16px"} alignItems={"center"}>
          <InputGroup
            size="md"
            width={"296px"}
            bg={"gray.650"}
            color={"gray.150"}
            variant={"filled"}
            borderRadius={"6px"}
          >
            <InputLeftElement pointerEvents="none">
              <Icon as={FiSearch} color="gray.100" />
            </InputLeftElement>
            <Input
              autoFocus
              ref={searchRef}
              value={searchState.rawSearchTerm}
              onChange={handleOnChange}
              placeholder={"Search"}
            />
            {searchState.type !== "init" && (
              <InputRightElement>
                <SmallCloseIcon onClick={handleClearSearch} />
              </InputRightElement>
            )}
          </InputGroup>
          <ButtonGroup>
            <Button
              size={"sm"}
              ml={2}
              onClick={handlePriorMatchClick}
              isDisabled={searchState.type !== "success" || searchState.searchMatchesIndices.length === 0}
              colorScheme={"darkBlue"}
              leftIcon={<MdArrowBackIosNew />}
            >
              Previous
            </Button>
            <Button
              size={"sm"}
              ml={2}
              onClick={handleNextMatchClick}
              isDisabled={searchState.type !== "success" || searchState.searchMatchesIndices.length === 0}
              colorScheme={"darkBlue"}
              rightIcon={<MdArrowForwardIos />}
            >
              Next
            </Button>
          </ButtonGroup>
          {searchState.rawSearchTerm.length > 0 && (
            <Flex ml={2} alignItems={"center"}>
              {searchState.type === "success" && (
                <Text
                  align={"left"}
                  color={searchState.searchMatchesIndices.length === 0 ? "red" : "kurtosisGreen.400"}
                >
                  {searchState.searchMatchesIndices.length > 0 && searchState.currentSearchIndex !== undefined && (
                    <span>
                      <Editable
                        display={"inline"}
                        p={0}
                        m={"0 4px 0 0"}
                        size={"sm"}
                        value={`${searchState.currentSearchIndex + 1}`}
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
                  <span>{searchState.searchMatchesIndices.length} matches</span>
                </Text>
              )}
            </Flex>
          )}
        </Flex>
        {searchState.type === "error" && <FormErrorMessage>{searchState.error}</FormErrorMessage>}
      </FormControl>
    );
  }
};
