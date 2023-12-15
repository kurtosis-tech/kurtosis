import { Text, TextProps, Tooltip } from "@chakra-ui/react";
import { DateTime } from "luxon";
import { useEffect, useState } from "react";
import { isDefined } from "./utils";

type FormatDateTimeProps = TextProps & {
  dateTime: DateTime | null;
  format: Intl.DateTimeFormatOptions | "relative";
};

export const FormatDateTime = ({ dateTime, format, ...textProps }: FormatDateTimeProps) => {
  const [formattedDateTime, setFormattedDateTime] = useState(
    format === "relative" ? dateTime?.toRelative() : dateTime?.toLocaleString(format),
  );

  useEffect(() => {
    if (format === "relative") {
      const timeout = setTimeout(() => {
        setFormattedDateTime(dateTime?.toRelative());
      }, 15 * 1000);
      return () => clearTimeout(timeout);
    }
  }, [dateTime, format]);

  if (!isDefined(dateTime)) {
    return (
      <Text as={"span"} {...textProps}>
        Unknown
      </Text>
    );
  }

  return (
    <Tooltip label={dateTime.toLocal().toFormat("yyyy-MM-dd HH:mm:ss ZZZZ")}>
      <Text as={"span"} {...textProps}>
        {formattedDateTime}
      </Text>
    </Tooltip>
  );
};
