import { ButtonGroup } from "@chakra-ui/react";
import { CodeEditor } from "./CodeEditor";
import { CopyButton } from "./CopyButton";
import { DownloadButton } from "./DownloadButton";
import { TitledCard } from "./TitledCard";

type FileDisplayProps = {
  title: string;
  value: string;
  filename: string;
};

export const FileDisplay = ({ value, filename, title }: FileDisplayProps) => {
  return (
    <TitledCard
      title={title}
      controls={
        <ButtonGroup>
          <CopyButton contentName={title.toLowerCase()} valueToCopy={value} isIconButton aria-label={`Copy ${title}`} />
          <DownloadButton
            fileName={filename}
            valueToDownload={value}
            isIconButton
            aria-label={`Download ${filename}`}
          />
        </ButtonGroup>
      }
    >
      <CodeEditor text={value} />
    </TitledCard>
  );
};
