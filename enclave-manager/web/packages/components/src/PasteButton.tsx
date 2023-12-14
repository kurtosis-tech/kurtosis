import { Button, ButtonProps, IconButton, IconButtonProps } from "@chakra-ui/react";
import { useState } from "react";
import { FiClipboard } from "react-icons/fi";
import { isDefined } from "./utils";

type PasteButtonProps<IsIconButton extends boolean> = (IsIconButton extends true ? IconButtonProps : ButtonProps) & {
  onValuePasted: (value: string) => void;
  text?: IsIconButton extends true ? string : never;
  isIconButton?: IsIconButton;
};

export const PasteButton = <IsIconButton extends boolean>({
  onValuePasted,
  text,
  isIconButton,
  ...buttonProps
}: PasteButtonProps<IsIconButton>) => {
  const [isLoading, setIsLoading] = useState(false);

  const handlePasteClick = async () => {
    setIsLoading(true);
    const value = await navigator.clipboard.readText();
    setIsLoading(false);
    if (isDefined(value)) {
      onValuePasted(value);
    }
  };

  // Firefox does not support programmatic clipboard.readText
  //https://developer.mozilla.org/en-US/docs/Web/API/Clipboard/readText
  if (!isDefined(onValuePasted) || !isDefined(navigator.clipboard.readText)) {
    return null;
  }

  if (isIconButton) {
    return (
      <IconButton
        icon={<FiClipboard />}
        size={"xs"}
        variant={"ghost"}
        colorScheme={"darkBlue"}
        onClick={handlePasteClick}
        isLoading={isLoading}
        {...(buttonProps as IconButtonProps)}
      >
        {text || "Paste"}
      </IconButton>
    );
  } else {
    return (
      <Button
        leftIcon={<FiClipboard />}
        size={"xs"}
        colorScheme={"darkBlue"}
        onClick={handlePasteClick}
        isLoading={isLoading}
        {...buttonProps}
      >
        {text || "Paste"}
      </Button>
    );
  }
};
