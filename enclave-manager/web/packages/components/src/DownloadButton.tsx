import { Button, ButtonProps, IconButton, IconButtonProps } from "@chakra-ui/react";
import { FiDownload } from "react-icons/fi";
import streamsaver from "streamsaver";
import { isAsyncIterable, isDefined, saveTextAsFile, stripAnsi } from "./utils";

type DownloadButtonProps<IsIconButton extends boolean> = (IsIconButton extends true ? IconButtonProps : ButtonProps) & {
  valueToDownload?: (() => string) | (() => AsyncIterable<string>) | string | null;
  fileName: string;
  text?: IsIconButton extends true ? string : never;
  isIconButton?: IsIconButton;
};

export const DownloadButton = <IsIconButton extends boolean>({
  valueToDownload,
  text,
  fileName,
  isIconButton,
  ...buttonProps
}: DownloadButtonProps<IsIconButton>) => {
  const handleDownloadClick = async () => {
    if (isDefined(valueToDownload)) {
      const v = typeof valueToDownload === "string" ? valueToDownload : valueToDownload();

      if (isAsyncIterable(v)) {
        const writableStream = streamsaver.createWriteStream(fileName);
        const writer = writableStream.getWriter();

        for await (const part of v) {
          await writer.write(new TextEncoder().encode(`${stripAnsi(part)}\n`));
        }
        await writer.close();
        return;
      }
      saveTextAsFile(v, fileName);
    }
  };

  if (!isDefined(valueToDownload) && !isDefined(buttonProps.onClick)) {
    return null;
  }

  if (isIconButton) {
    return (
      <IconButton
        icon={<FiDownload />}
        size={"xs"}
        variant={"ghost"}
        colorScheme={"darkBlue"}
        onClick={handleDownloadClick}
        {...(buttonProps as IconButtonProps)}
      />
    );
  } else {
    return (
      <Button
        leftIcon={<FiDownload />}
        size={"xs"}
        colorScheme={"darkBlue"}
        onClick={handleDownloadClick}
        {...buttonProps}
      >
        {text || "Download"}
      </Button>
    );
  }
};
