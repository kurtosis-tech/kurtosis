import { Box, Button, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { FiDownloadCloud } from "react-icons/fi";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";

export const KurtosisUpgrader = () => {
  const kurtosisClient = useKurtosisClient();
  const [isNewKurtosisVersionAvailable, setIsNewKurtosisVersionAvailable] = useState(false);
  const [latestKurtosisVersion, setLatestKurtosisVersion] = useState("");

  useEffect(() => {
    //if (kurtosisClient.isRunningInCloud()) {
    checkForNewKurtosisVersion(isNewKurtosisVersionAvailable);
    //}
  });

  const checkForNewKurtosisVersion = async (isNewKurtosisVersionAvailable: boolean) => {
    try {
      const isNewKurtosisVersionAvailableResponse = await kurtosisClient.isNewKurtosisVersionAvailable();
      setIsNewKurtosisVersionAvailable(isNewKurtosisVersionAvailableResponse.isAvailable);
      setLatestKurtosisVersion(isNewKurtosisVersionAvailableResponse.latestVersion);
      //TODO remove these console logs
      console.log("LATEST KURTOSIS VERSION:");
      console.log(latestKurtosisVersion);
      console.log("IS NEW KURTOSIS VERSION AVAILABLE:");
      console.log(isNewKurtosisVersionAvailable);
    } catch (error) {
      console.error(`Error occurred when checking for new Kurtosis version. ${error}`);
    }
  };

  const upgradeKurtosis = async () => {
    console.log("upgrading Kurtosis...");
    try {
      const upgradeKurtosisVersionResponse = await kurtosisClient.upgradeKurtosisVersion();
      console.log("...Kurtosis successfully upgraded");
    } catch (error) {
      console.error(`Error occurred when upgrading Kurtosis to the latest version. ${error}`);
    }
  };

  if (!isNewKurtosisVersionAvailable) {
    return null;
  }
  return (
    <Box borderWidth="1px" borderRadius="lg" borderColor="red">
      <Text>A new Kurtosis version is available {latestKurtosisVersion}</Text>
      <Button colorScheme={"green"} leftIcon={<FiDownloadCloud />} size={"sm"} onClick={upgradeKurtosis}>
        Upgrade Kurtosis
      </Button>
    </Box>
  );
};
