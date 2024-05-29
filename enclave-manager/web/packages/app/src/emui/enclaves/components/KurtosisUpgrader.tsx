import { Alert, AlertDescription, AlertIcon, Box, Button, Flex } from "@chakra-ui/react";
import { GetCloudInstanceConfigResponse } from "enclave-manager-sdk/build/kurtosis_backend_server_api_pb";
import { PropsWithChildren, useCallback, useEffect, useState } from "react";
import { FiDownloadCloud, FiRotateCcw } from "react-icons/fi";
import { GoBug } from "react-icons/go";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { GITHUB_ISSUE_URL } from "../../constants";

const UpgradeAlert = ({
  children,
  status,
}: PropsWithChildren<{ status: "info" | "warning" | "error" | "success" }>) => {
  return (
    <Alert status={status}>
      <AlertIcon />
      <AlertDescription width={"100%"}>
        <Flex justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
          {children}
        </Flex>
      </AlertDescription>
    </Alert>
  );
};

enum UpgradeStatus {
  NONE, // Default state
  AVAILABLE,
  IN_PROGRESS,
  SUCCESS,
  ERROR,
}

const skipCache = true;

export const KurtosisUpgrader = () => {
  const kurtosisClient = useKurtosisClient();
  const [upgradeStatus, setUpgradeStatus] = useState<UpgradeStatus>(UpgradeStatus.NONE);
  const [latestKurtosisVersion, setLatestKurtosisVersion] = useState("");

  const checkForNewKurtosisVersion = async () => {
    try {
      const isNewKurtosisVersionAvailableResponse = await kurtosisClient.isNewKurtosisVersionAvailable();
      if (isNewKurtosisVersionAvailableResponse.isAvailable) {
        setUpgradeStatus(UpgradeStatus.AVAILABLE);
        setLatestKurtosisVersion(isNewKurtosisVersionAvailableResponse.latestVersion);
      }
    } catch (error) {
      console.error(`Error occurred when checking for new Kurtosis version. ${error}`);
    }
  };

  const upgradeKurtosis = async () => {
    try {
      setUpgradeStatus(UpgradeStatus.IN_PROGRESS);
      await kurtosisClient.upgradeKurtosisVersion();
    } catch (error) {
      setUpgradeStatus(UpgradeStatus.ERROR);
      console.error(`Error occurred while upgrading Kurtosis to the latest version. ${error}`);
    }
  };

  const wait = (ms: number) => {
    return new Promise((resolve) => setTimeout(resolve, ms));
  };

  const getCloudInstanceConfigWithRetry = useCallback(
    async (tries: number, interval: number): Promise<GetCloudInstanceConfigResponse> => {
      let getCloudInstanceConfigResponse: GetCloudInstanceConfigResponse = new GetCloudInstanceConfigResponse();
      try {
        getCloudInstanceConfigResponse = await kurtosisClient.getCloudInstanceConfig(skipCache);
      } catch (e) {
        const newTries = tries - 1;

        if (newTries === 0) {
          throw e;
        }

        await wait(interval);

        return getCloudInstanceConfigWithRetry(newTries, interval);
      }
      return getCloudInstanceConfigResponse;
    },
    [],
  );

  // Check once on load if a new Kurtosis version is available
  useEffect(() => {
    if (!kurtosisClient.isRunningInCloud()) return;

    checkForNewKurtosisVersion();

    // Only run effect once, ignore eslint warning
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // If upgrade is in progress, check the status every 2 seconds
  useEffect(() => {
    if (upgradeStatus !== UpgradeStatus.IN_PROGRESS) return;

    const interval = setInterval(async () => {
      try {
        // Calling it with retries because when the engine is restarted it won't return a response for a few seconds
        const getCloudInstanceConfigResponse = await getCloudInstanceConfigWithRetry(5, 2000);
        const instanceStatus = getCloudInstanceConfigResponse.status;
        if (instanceStatus === "running") {
          setUpgradeStatus(UpgradeStatus.SUCCESS);
          clearInterval(interval);
        }
      } catch (error) {
        console.error(`Error occurred getting the cloud instance config. ${error}`);
        setUpgradeStatus(UpgradeStatus.ERROR);
        clearInterval(interval);
      }
    }, 10000);

    return () => clearInterval(interval);
  }, [upgradeStatus, kurtosisClient, getCloudInstanceConfigWithRetry]);

  if (upgradeStatus === UpgradeStatus.NONE) {
    return null;
  }

  return (
    <Box width={"100%"}>
      {upgradeStatus === UpgradeStatus.AVAILABLE && (
        <UpgradeAlert status="warning">
          A new Kurtosis version (v{latestKurtosisVersion}) is available.
          <Button colorScheme={"orange"} leftIcon={<FiDownloadCloud />} size={"sm"} onClick={upgradeKurtosis}>
            Upgrade Cloud Instance
          </Button>
        </UpgradeAlert>
      )}

      {upgradeStatus === UpgradeStatus.IN_PROGRESS && (
        <UpgradeAlert status="info">
          Upgrading Kurtosis to version: v{latestKurtosisVersion}
          <Button colorScheme={"blue"} isLoading size={"sm"}>
            Loading
          </Button>
        </UpgradeAlert>
      )}

      {upgradeStatus === UpgradeStatus.SUCCESS && (
        <UpgradeAlert status="success">
          Kurtosis has been upgraded to version: v{latestKurtosisVersion}. Please refresh the page.
          <Button
            colorScheme={"green"}
            leftIcon={<FiRotateCcw />}
            size={"sm"}
            onClick={() => {
              window.location.reload();
            }}
          >
            Refresh
          </Button>
        </UpgradeAlert>
      )}

      {upgradeStatus === UpgradeStatus.ERROR && (
        <UpgradeAlert status="error">
          Upgrading Kurtosis version failed. Please open a GitHub issue for support.
          <Button
            colorScheme={"red"}
            as={"a"}
            href={`${GITHUB_ISSUE_URL}&version=${latestKurtosisVersion}`}
            leftIcon={<GoBug />}
            target={"_blank"}
            size={"sm"}
          >
            Report a Bug
          </Button>
        </UpgradeAlert>
      )}
    </Box>
  );
};
