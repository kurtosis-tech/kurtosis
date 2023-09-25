import React, { useCallback, useEffect, useRef, useState } from "react";
import { VariableSizeList as List } from "react-window";
import useWindowDimensions from "../../utils/windowDimension";
import { Flex, Box, Spacer } from "@chakra-ui/react";
import { TriangleDownIcon, DownloadIcon } from '@chakra-ui/icons'
import parse from 'html-react-parser';
import hasAnsi from 'has-ansi';
import stripAnsi from 'strip-ansi';
import { useClipboard } from '@chakra-ui/react'
import { saveTextAsFile } from "../../utils/download";

const Convert = require('ansi-to-html');
const convert = new Convert();

const Row = ({ data, index, setSize, windowWidth}) => {
  const rowRef = useRef();
  
  useEffect(() => {
    if (rowRef.current !== undefined) {
      setSize(index, rowRef.current.getBoundingClientRect().height);
    } 
  }, [setSize, index, windowWidth]);

  if (data[index] !== undefined && data[index].length !== 0) {
    let txt = data[index]
    if (hasAnsi(data[index])) {
      const parsedAnsi = convert.toHtml(data[index])
      txt = parse(parsedAnsi)
    }
    return (
        <div
          ref={rowRef}
          style={styles.row}
        >
         {txt}
        </div>
      );
  } 
  return null;
};

export const Log = ({logs, fileName}) => {
  const { onCopy, value:copyValue, setValue:setCopyValue, hasCopied } = useClipboard("");

  const listRef = useRef(null);
  const sizeMap = useRef({});
  const testRef = useRef({})
  const containerRef = useRef({})
  const [previousOffset, setPreviousOffset] = useState(0)

  const setSize = useCallback((index, size) => {
    sizeMap.current = { ...sizeMap.current, [index]: size };
    listRef.current.resetAfterIndex(index);
  }, []);

  const getSize = index => sizeMap.current[index] || 50;
  const {width: windowWidth, height: windowHeight} = useWindowDimensions();

  const [isBottom, setBottom] = useState(true)

  useEffect(() => {
    const logsWithoutAnsi = logs.map((log)=> {
      return stripAnsi(log)
    })
    setCopyValue(logsWithoutAnsi.join("\n"))    
    // automatically scroll the bottom if it;s at the bottom
    if (logs.length > 0 && isBottom) {
      scrollToBottom()
    }
  }, [logs]);
  
  const scrollToBottom = () => {
    if (listRef.current) {
      //console.log("scroll to ", logs.length)
      listRef.current.scrollToItem(logs.length - 1, "end");
    }
  }

  const handleScroll = ({scrollOffset}) => {
      // user manually scrolled up 
      if (scrollOffset < previousOffset && testRef.current.clientHeight > containerRef.current.clientHeight) {
        //console.log("setting scroll to false!")
        setBottom(false);
      } 
      setPreviousOffset(Math.floor(scrollOffset))
  };

  const handleToBottom = () => {
    scrollToBottom()    
    setBottom(true)
  }

  const handleItemsRendered = ({
    overscanStartIndex,
    overscanStopIndex,
    visibleStartIndex,
    visibleStopIndex}) => {
      const maybeBottom = [overscanStopIndex, visibleStopIndex]
      if (overscanStopIndex === logs.length - 1) {
        setBottom(true)
      }
  }

  const handleDownload = () => {
    const logsWithoutAnsi = logs.map((log)=> {
      return stripAnsi(log)
    })
    const logsAsText = logsWithoutAnsi.join("\n")
    saveTextAsFile(logsAsText, fileName)
  }

  return (
    <div className="flex flex-col bg-black">
      <List
        ref={listRef}
        height={(windowHeight-100)*0.8}
        width={"100%"}
        itemCount={logs.length}
        itemSize={getSize}
        itemData={logs}
        onScroll={handleScroll}
        outerRef={containerRef}
        innerRef={testRef}
        style={{backgroundColor: "black"}}
        onItemsRendered={handleItemsRendered}
      >
      {({ data, index, style }) => (
        <div style={style}>
          <Row
            data={data}
            index={index}
            setSize={setSize}
            windowWidth={windowWidth}
            style={style}
          />
        </div>
      )}
      </List>
      <Flex className="bg-black" style={{height: `80px`}}>
        <Spacer />
        <Box p='2' m="4" onClick={onCopy} className="rounded-md border-white border-1" height={"40px"}> 
          <svg class="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 20">
            <path d="M5 9V4.13a2.96 2.96 0 0 0-1.293.749L.879 7.707A2.96 2.96 0 0 0 .13 9H5Zm11.066-9H9.829a2.98 2.98 0 0 0-2.122.879L7 1.584A.987.987 0 0 0 6.766 2h4.3A3.972 3.972 0 0 1 15 6v10h1.066A1.97 1.97 0 0 0 18 14V2a1.97 1.97 0 0 0-1.934-2Z"/>
            <path d="M11.066 4H7v5a2 2 0 0 1-2 2H0v7a1.969 1.969 0 0 0 1.933 2h9.133A1.97 1.97 0 0 0 13 18V6a1.97 1.97 0 0 0-1.934-2Z"/>
          </svg>
        </Box>
        <Box p='2' m="4" onClick={handleDownload} className="rounded-md border-white border-1" height={"40px"}> 
          <DownloadIcon textColor={"white"}/>
        </Box>
        <Box p='2' m="4" onClick={handleToBottom} className="rounded-md border-white border-1" height={"40px"}> 
          <TriangleDownIcon textColor={"white"}/>
        </Box>
      </Flex>
    </div>
  );
}

const styles = {
  row: {
    fontFamily: "system-ui",
    boxSizing: "border-box",
    borderBottom: "1px solid #222",
    padding: "1em",
    color: "white"
  }
};