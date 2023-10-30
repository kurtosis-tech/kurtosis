import { Button, ButtonProps, useToast } from "@chakra-ui/react";
import { FiCopy } from "react-icons/fi";
import { isDefined } from "../utils";

type CopyButtonProps = ButtonProps & {
  valueToCopy?: string | null;
};

export const CopyButton = ({ valueToCopy, ...buttonProps }: CopyButtonProps) => {
  const toast = useToast();

  const handleCopyClick = () => {
    if (isDefined(valueToCopy)) {
      navigator.clipboard.writeText(valueToCopy);
      toast({
        title: `Copied '${valueToCopy}' to the clipboard`,
        status: `success`,
      });
    }
  };

  if (!isDefined(valueToCopy)) {
    return null;
  }

  return (
    <Button leftIcon={<FiCopy />} size={"xs"} colorScheme={"darkBlue"} onClick={handleCopyClick} {...buttonProps}>
      Copy
    </Button>
  );
};
