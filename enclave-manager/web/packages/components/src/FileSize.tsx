import { Text, TextProps } from "@chakra-ui/react";
import { isDefined } from "./utils";

type FileSizeProps = TextProps & {
  fileSize?: bigint;
};

const units = ["B", "KB", "MB", "GB", "TB"];

export const FileSize = ({ fileSize, ...textProps }: FileSizeProps) => {
  if (!isDefined(fileSize)) {
    return null;
  }
  let size = fileSize;
  let unitIndex = 0;
  while (size > 1024 && unitIndex < units.length - 1) {
    size = size / BigInt(1024);
    unitIndex += 1;
  }

  return (
    <Text as={"span"} {...textProps}>
      {`${size}${units[unitIndex]}`}
    </Text>
  );
};
