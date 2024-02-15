import { Button, ButtonProps } from "@chakra-ui/react";
import { BiPaintRoll } from "react-icons/bi";

export const FormatButton = ({ ...buttonProps }: ButtonProps) => {
  return (
    <Button
      leftIcon={<BiPaintRoll />}
      size={"sm"}
      colorScheme={"darkBlue"}
      loadingText={"Format"}
      variant={"outline"}
      {...buttonProps}
    >
      Format
    </Button>
  );
};
