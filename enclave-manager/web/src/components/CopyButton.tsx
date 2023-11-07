import { Button, ButtonProps, useToast } from "@chakra-ui/react";
import { FiCopy } from "react-icons/fi";
import { isDefined } from "../utils";

type CopyButtonProps = ButtonProps & {
  valueToCopy?: (() => string) | string | null;
  text?: string;
  contentName: string;
};

export const CopyButton = ({ valueToCopy, text, contentName, ...buttonProps }: CopyButtonProps) => {
  const toast = useToast();

  const handleCopyClick = () => {
    if (isDefined(valueToCopy)) {
      const v = typeof valueToCopy === "string" ? valueToCopy : valueToCopy();
      navigator.clipboard.writeText(v);
      toast({
        title: `Copied ${contentName} to the clipboard`,
        status: `success`,
      });
    }
  };

  if (!isDefined(valueToCopy)) {
    return null;
  }

  return (
    <Button leftIcon={<FiCopy />} size={"xs"} colorScheme={"darkBlue"} onClick={handleCopyClick} {...buttonProps}>
      {text || "Copy"}
    </Button>
  );
};
