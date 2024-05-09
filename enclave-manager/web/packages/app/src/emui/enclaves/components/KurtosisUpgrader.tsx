import { Box, Button, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { FiDownloadCloud } from "react-icons/fi";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";

export const KurtosisUpgrader = () => {
  const kurtosisClient = useKurtosisClient();
  const skipCache = true;
  const [isNewKurtosisVersionAvailable, setIsNewKurtosisVersionAvailable] = useState(false);
  const [latestKurtosisVersion, setLatestKurtosisVersion] = useState("");
  const [isUpgradeInProgress, setIsUpgradeInProgress] = useState(false);
  const [isUpgradeFinished, setIsUpgradeFinished] = useState(false);

  //TODO add error messages when something fails
  //TODO add the refresh page button when the upgrade finished
  //TODO add upgrade in progress icon or animation

  useEffect(() => {
    //if (kurtosisClient.isRunningInCloud()) {
    checkForNewKurtosisVersion(isNewKurtosisVersionAvailable);
    //Implementing the setInterval method
    const interval = setInterval(() => {
      if (isUpgradeInProgress) {
        try {
          kurtosisClient.getCloudInstanceConfig(skipCache).then((getCloudInstanceConfigResponse) => {
            const instanceStatus = getCloudInstanceConfigResponse.status;
            console.log(`Instance status in interval: ${instanceStatus}`);
            if (instanceStatus === "running") {
              setIsUpgradeInProgress(false);
              setIsUpgradeFinished(true);
              clearInterval(interval);
            }
          });
        } catch (error) {
          console.error(`Error occurred getting the cloud instance config. ${error}`);
        }
      }
      if (isUpgradeFinished) {
        clearInterval(interval);
      }
    }, 2000);

    //Clearing the interval
    return () => clearInterval(interval);
    //}
  });

  const checkForNewKurtosisVersion = async (isNewKurtosisVersionAvailable: boolean) => {
    try {
      const isNewKurtosisVersionAvailableResponse = await kurtosisClient.isNewKurtosisVersionAvailable();
      setIsNewKurtosisVersionAvailable(isNewKurtosisVersionAvailableResponse.isAvailable);
      setLatestKurtosisVersion(isNewKurtosisVersionAvailableResponse.latestVersion);
      //TODO remove these console logs
      console.log(`LATEST KURTOSIS VERSION: ${latestKurtosisVersion}`);
      console.log(`IS NEW KURTOSIS VERSION AVAILABLE: ${isNewKurtosisVersionAvailable}`);
    } catch (error) {
      console.error(`Error occurred when checking for new Kurtosis version. ${error}`);
    }
  };

  const upgradeKurtosis = async () => {
    console.log("User pressed the upgrade button");
    try {
      const upgradeKurtosisVersionResponse = await kurtosisClient.upgradeKurtosisVersion();
      const getCloudInstanceConfigResponse = await kurtosisClient.getCloudInstanceConfig(skipCache);
      const instanceStatus = getCloudInstanceConfigResponse.status;
      if (instanceStatus === "upgrading") {
        console.log(`Instance status in upgradeKurtosis: ${instanceStatus}`);
      } else {
        console.log(`Instance status in upgradeKurtosis: ${instanceStatus}`);
      }
      setIsUpgradeInProgress(true);
    } catch (error) {
      setIsUpgradeInProgress(false);
      console.error(`Error occurred when upgrading Kurtosis to the latest version. ${error}`);
    }
  };

  if (!isNewKurtosisVersionAvailable) {
    return null;
  }
  return (
    <Box>
      {!isUpgradeFinished && !isUpgradeInProgress && (
        <Text fontSize="xs">A new Kurtosis version is available {latestKurtosisVersion}</Text>
      )}
      {!isUpgradeFinished && !isUpgradeInProgress && (
        <Button colorScheme={"green"} leftIcon={<FiDownloadCloud />} size={"sm"} onClick={upgradeKurtosis}>
          Upgrade Kurtosis
        </Button>
      )}
      {isUpgradeInProgress && <Text fontSize="xs">Upgrading Kurtosis to version {latestKurtosisVersion}...</Text>}
      {isUpgradeFinished && <Text fontSize="xs">Kurtosis has been updated to version {latestKurtosisVersion}</Text>}
    </Box>
  );
};
