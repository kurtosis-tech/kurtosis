import React, {useEffect, useRef, useState} from "react";
import useWindowDimensions from "../../utils/windowDimension";
import {Box, Flex, Spacer, Text, Tooltip, useClipboard} from "@chakra-ui/react";
import parse from 'html-react-parser';
import hasAnsi from 'has-ansi';
import stripAnsi from 'strip-ansi';
import {saveTextAsFile} from "../../utils/download";
import {Virtuoso} from "react-virtuoso";
import {DownloadIcon, TriangleDownIcon} from '@chakra-ui/icons'

var os = require('os-browserify/browser')
const Convert = require('ansi-to-html');
const convert = new Convert();

const Row = ({log}) => {
    if (log !== undefined && log.length !== 0) {
        let txt = log
        if (hasAnsi(txt)) {
            const parsedAnsi = convert.toHtml(txt)
            txt = parse(parsedAnsi)
        }
        return (
            <Box style={styles.row}>{txt}</Box>
        );
    }
    return <></>;
};

export const Log = ({logs, fileName, currentExecutionStatus, executionStatusText}) => {
    const [displayLogs, setDisplayLogs] = useState(logs)
    const virtuosoRef = useRef(null)
    const {height: windowHeight} = useWindowDimensions();
    const {onCopy, setValue: setCopyValue, hasCopied} = useClipboard("");
    const [logsExecutionStatus, setLogsExecutionStatus] = useState(<></>)

    useEffect(() => {
        setDisplayLogs(logs);
        const logsWithoutAnsi = logs.map((log) => {
            return stripAnsi(log)
        })
        setCopyValue(logsWithoutAnsi.join(os.EOL))
        setLogsExecutionStatus(currentExecutionStatus)
    }, [logs]);

    const handleDownload = () => {
        const logsWithoutAnsi = logs.map((log) => {
            return stripAnsi(log)
        })
        const logsAsText = logsWithoutAnsi.join(os.EOL)
        saveTextAsFile(logsAsText, fileName)
    }

    const handleScrollToBottom = () => {
        virtuosoRef.current.scrollToIndex({
            index: displayLogs.length - 1, behavior: 'smooth', align: "end"
        });
    }

    return (
        <div className="flex flex-col bg-black">
            <Virtuoso
                style={{height: (windowHeight - 100) * 0.8, backgroundColor: "black"}}
                ref={virtuosoRef}
                data={displayLogs}
                totalCount={displayLogs.length - 1}
                itemContent={(index, log) => {
                    return (<Row index={index} log={log}/>);
                }}
                followOutput={"smooth"}
            />
            <Flex className="bg-black" style={{height: `80px`}}>
                <Spacer/>
                <Box
                    p='2'
                    m="4"
                    height={"40px"}
                >
                    <Text fontSize={"sm"}>{executionStatusText}</Text>
                </Box>
                <Box p='2' m="4"
                     height={"40px"}
                >
                    {logsExecutionStatus}
                </Box>
                <Tooltip label={`${hasCopied ? "Copied!" : "Copy to clipboard"}`}
                         placement='top-end'
                         closeOnClick={false}
                >
                    <Box p='2' m="4" onClick={onCopy} height={"40px"}>
                        <svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true"
                             xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 20">
                            <path
                                d="M5 9V4.13a2.96 2.96 0 0 0-1.293.749L.879 7.707A2.96 2.96 0 0 0 .13 9H5Zm11.066-9H9.829a2.98 2.98 0 0 0-2.122.879L7 1.584A.987.987 0 0 0 6.766 2h4.3A3.972 3.972 0 0 1 15 6v10h1.066A1.97 1.97 0 0 0 18 14V2a1.97 1.97 0 0 0-1.934-2Z"/>
                            <path
                                d="M11.066 4H7v5a2 2 0 0 1-2 2H0v7a1.969 1.969 0 0 0 1.933 2h9.133A1.97 1.97 0 0 0 13 18V6a1.97 1.97 0 0 0-1.934-2Z"/>
                        </svg>
                    </Box>
                </Tooltip>
                <Tooltip label={"Download to file"}
                         placement='top'
                >
                    <Box p='2' m="4" onClick={handleDownload} height={"40px"}>
                        <DownloadIcon textColor={"white"}/>
                    </Box>
                </Tooltip>
                <Tooltip label={"Autoscroll to bottom"}
                         placement='top'
                >
                    <Box p='2' m="4"
                         height={"40px"}
                         onClick={handleScrollToBottom}
                    >
                        <TriangleDownIcon textColor={"white"}/>
                    </Box>
                </Tooltip>
            </Flex>
        </div>
    )
}

const styles = {
    row: {
        fontSize: "10pt",
        fontFamily: "Menlo, Monaco, Inconsolata, Consolas, Courier, monospace",
        boxSizing: "border-box",
        borderBottom: "1px solid #222",
        padding: "1em",
        color: "white"
    }
};
