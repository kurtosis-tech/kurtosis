import { ButtonGroup, Card, Flex, Text } from "@chakra-ui/react";
import { CodeEditor } from "./CodeEditor";
import { CopyButton } from "./CopyButton";
import { DownloadButton } from "./DownloadButton";

type FileDisplayProps = {
  title: string;
  value: string;
  filename: string;
};

export const FileDisplay = ({ value, filename, title }: FileDisplayProps) => {
  return (
    <Flex flexDirection={"column"} gap={"12px"} height={"100%"}>
      <Flex justifyContent={"space-between"}>
        <Text fontSize={"sm"} fontWeight={"medium"}>
          {title}
        </Text>
        <ButtonGroup isAttached>
          <CopyButton contentName={title.toLowerCase()} valueToCopy={value} />
          <DownloadButton fileName={filename} valueToDownload={value} />
        </ButtonGroup>
      </Flex>
      <Card height={"100%"}>
        <CodeEditor text={value} />
      </Card>
    </Flex>
  );
};
