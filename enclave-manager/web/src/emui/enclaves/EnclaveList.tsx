import {
  Alert,
  AlertDescription,
  AlertIcon,
  AlertTitle,
  Button,
  ButtonGroup,
  Flex,
  Spinner,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { FiPlus, FiTrash2 } from "react-icons/fi";
import { ResultNS } from "true-myth";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { EnclavesTable } from "../../components/enclaves/EnclavesTable";
import { isDefined } from "../../utils";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const [error, setError] = useState<string>();
  const kurtosisClient = useKurtosisClient();
  const [enclaves, setEnclaves] = useState<EnclaveFullInfo[]>();
  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  useEffect(() => {
    setError(undefined);
    (async () => {
      const enclavesResponse = await kurtosisClient.getEnclaves();
      if (enclavesResponse.isErr) {
        setError(enclavesResponse.error.message);
        return;
      }
      const enclaves = Object.values(enclavesResponse.value.enclaveInfo);
      const [starlarkRuns, services, filesAndArtifacts] = await Promise.all([
        Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
        Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
        Promise.all(enclaves.map((enclave) => kurtosisClient.listFilesArtifactNamesAndUuids(enclave))),
      ]);

      const starlarkErrors = starlarkRuns.filter(ResultNS.isErr);
      const servicesErrors = services.filter(ResultNS.isErr);
      const filesAndArtifactErrors = filesAndArtifacts.filter(ResultNS.isErr);
      if (starlarkErrors.length + servicesErrors.length + filesAndArtifactErrors.length > 0) {
        setError(
          `Starlark errors: ${
            starlarkErrors.length > 0 ? starlarkErrors.map((r) => r.error.message).join("\n") : "None"
          }\nServices errors: ${
            servicesErrors.length > 0 ? servicesErrors.map((r) => r.error.message).join("\n") : "None"
          }\nFiles and Artifacts errors: ${
            filesAndArtifactErrors.length > 0 ? filesAndArtifactErrors.map((r) => r.error.message).join("\n") : "None"
          }`,
        );
        return;
      }

      setEnclaves(
        enclaves.map((enclave, i) => ({
          ...enclave,
          // These values are never actually null because of the checking above
          starlarkRun: starlarkRuns[i].unwrapOr(null)!,
          services: services[i].unwrapOr(null)!,
          filesAndArtifacts: filesAndArtifacts[i].unwrapOr(null)!,
        })),
      );
    })();
  }, [kurtosisClient]);

  return (
    <Flex direction="column">
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosisGreen"}>
        <Flex justifyContent={"space-between"}>
          <TabList>
            <Tab>Enclaves</Tab>
          </TabList>
          <Flex gap={"24px"} alignItems={"center"}>
            <ButtonGroup isAttached variant={"kurtosisGroupOutline"} size={"sm"}>
              <Button colorScheme={"blue"} onClick={() => setSelectedEnclaves([])}>
                {selectedEnclaves.length} selected
              </Button>
              <Button colorScheme={"red"} leftIcon={<FiTrash2 />}>
                Delete
              </Button>
            </ButtonGroup>
            <Button variant={"kurtosisOutline"} colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"sm"}>
              Create Enclave
            </Button>
          </Flex>
        </Flex>
        <TabPanels>
          <TabPanel>
            {isDefined(enclaves) && (
              <EnclavesTable
                enclavesData={enclaves}
                selection={selectedEnclaves}
                onSelectionChange={setSelectedEnclaves}
              />
            )}
            {isDefined(error) && (
              <Alert status="error">
                <AlertIcon />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            {!isDefined(enclaves) && !isDefined(error) && (
              <Flex justifyContent={"center"} p={"20px"}>
                <Spinner size={"xl"} />
              </Flex>
            )}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
