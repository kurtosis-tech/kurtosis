import { Button, ButtonProps, IconButton, IconButtonProps, useToast } from "@chakra-ui/react";
import { FiCopy } from "react-icons/fi";
import { SuccessToast } from "./Toasts";
import { isDefined } from "./utils";

type CopyButtonProps<IsIconButton extends boolean> = (IsIconButton extends true ? IconButtonProps : ButtonProps) & {
  valueToCopy?: (() => string) | string | null;
  text?: IsIconButton extends true ? string : never;
  isIconButton?: IsIconButton;
  contentName: string;
};

export const CopyButton = <IsIconButton extends boolean>({
  valueToCopy,
  text,
  contentName,
  isIconButton,
  ...buttonProps
}: CopyButtonProps<IsIconButton>) => {
  const toast = useToast();

  const handleCopyClick = () => {
    if (isDefined(valueToCopy)) {
      const v = typeof valueToCopy === "string" ? valueToCopy : valueToCopy();
      navigator.clipboard.writeText(v);
      toast({
        position: "bottom",
        render: () => <SuccessToast message={`Copied ${contentName} to the clipboard`} />,
      });
    }
  };

  if (!isDefined(valueToCopy) && !isDefined(buttonProps.onClick)) {
    return null;
  }

  if (isIconButton) {
    return (
      <IconButton
        icon={<FiCopy />}
        size={"xs"}
        variant={"ghost"}
        colorScheme={"darkBlue"}
        onClick={handleCopyClick}
        {...(buttonProps as IconButtonProps)}
      >
        {text || "Copy"}
      </IconButton>
    );
  } else {
    return (
      <Button leftIcon={<FiCopy />} size={"xs"} colorScheme={"darkBlue"} onClick={handleCopyClick} {...buttonProps}>
        {text || "Copy"}
      </Button>
    );
  }
};
