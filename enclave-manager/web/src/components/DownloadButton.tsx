import { Button, ButtonProps, IconButton, IconButtonProps } from "@chakra-ui/react";
import { FiDownload } from "react-icons/fi";
import { isDefined } from "../utils";
import { saveTextAsFile } from "../utils/download";

type DownloadButtonProps<IsIconButton extends boolean> = (IsIconButton extends true ? IconButtonProps : ButtonProps) & {
  valueToDownload?: (() => string) | string | null;
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
  const handleDownloadClick = () => {
    if (isDefined(valueToDownload)) {
      const v = typeof valueToDownload === "string" ? valueToDownload : valueToDownload();
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
