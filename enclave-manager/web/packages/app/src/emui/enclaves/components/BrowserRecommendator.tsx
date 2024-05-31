import { Alert, AlertDescription, AlertIcon, Box } from "@chakra-ui/react";
import { isChrome } from "react-device-detect";

export const BrowserRecommendator = () => {
  if (isChrome) {
    return null;
  }

  return (
    <Box width={"100%"}>
      <Alert status="warning">
        <AlertIcon />
        <AlertDescription width={"100%"}>
          We recommend using Kurtosis Cloud with Google Chrome otherwise your experience may be degraded.
        </AlertDescription>
      </Alert>
    </Box>
  );
};
