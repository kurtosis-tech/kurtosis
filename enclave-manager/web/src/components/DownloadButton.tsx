import { Button, ButtonProps } from "@chakra-ui/react";
import { FiDownload } from "react-icons/fi";
import { isDefined } from "../utils";
import { saveTextAsFile } from "../utils/download";

type DownloadButtonProps = ButtonProps & {
  valueToDownload?: (() => string) | string | null;
  fileName: string;
  text?: string;
};

export const DownloadButton = ({ valueToDownload, text, fileName, ...buttonProps }: DownloadButtonProps) => {
  const handleDownloadClick = () => {
    if (isDefined(valueToDownload)) {
      const v = typeof valueToDownload === "string" ? valueToDownload : valueToDownload();
      saveTextAsFile(v, fileName);
    }
  };

  if (!isDefined(valueToDownload)) {
    return null;
  }

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
};
