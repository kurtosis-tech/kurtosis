import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Box, Link, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import {
  KURTOSIS_CLOUD_INSTANCE_MAX_UPTIME_IN_HOURS,
  KURTOSIS_CLOUD_SUBSCRIPTION_URL,
} from "../../../client/constants";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";

export const InstanceTerminationWarning = () => {
  const [cloudInstanceRemainingHours, setCloudInstanceRemainingHours] = useState(0);
  const kurtosisClient = useKurtosisClient();

  useEffect(() => {
    if (kurtosisClient.isRunningInCloud()) {
      fetchCloudInstanceCreationTime();
    }
  });

  const fetchCloudInstanceCreationTime = async () => {
    try {
      const cloudInstanceConfigResponse = await kurtosisClient.getCloudInstanceConfig();
      const upTime = Math.floor((Date.now() - new Date(cloudInstanceConfigResponse.created).getTime()) / (3600 * 1000));
      const remainingHours = KURTOSIS_CLOUD_INSTANCE_MAX_UPTIME_IN_HOURS - upTime - 1;
      setCloudInstanceRemainingHours(remainingHours);
    } catch (error) {
      console.error(error);
    }
  };

  if (cloudInstanceRemainingHours <= 0) {
    return null;
  }
  return (
    <Box borderWidth="1px" borderRadius="lg" borderColor="red" p={1}>
      <Text fontSize="xs">
        Your cloud instance will terminate in {cloudInstanceRemainingHours} hour(s) if you do not have a{" "}
        <Link href={`${KURTOSIS_CLOUD_SUBSCRIPTION_URL}`} isExternal>
          subscription <ExternalLinkIcon mx="1px" />
        </Link>
      </Text>
    </Box>
  );
};
