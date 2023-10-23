import { DateTime } from "luxon";
import { Text, TextProps, Tooltip } from "@chakra-ui/react";
import { isDefined } from "../utils";
import { useEffect, useState } from "react";

type RelativeDateTimeProps = TextProps & {
  dateTime: DateTime | null;
};

export const RelativeDateTime = ({ dateTime, ...textProps }: RelativeDateTimeProps) => {
  const [relativeTime, setRelativeTime] = useState(dateTime?.toRelative());

  useEffect(() => {
    const timeout = setTimeout(() => {
      setRelativeTime(dateTime?.toRelative());
    }, 15 * 1000);
    return () => clearTimeout(timeout);
  }, [dateTime]);

  if (!isDefined(dateTime)) {
    return (
      <Text as={"span"} {...textProps}>
        Unknown
      </Text>
    );
  }

  return (
    <Tooltip label={dateTime.toISO()}>
      <Text as={"span"} {...textProps}>
        {relativeTime}
      </Text>
    </Tooltip>
  );
};
