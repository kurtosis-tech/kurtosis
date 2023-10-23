import {
  Button,
  ButtonGroup,
  Card,
  Flex,
  Icon,
  Spinner,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
} from "@chakra-ui/react";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { useEffect, useState } from "react";
import { isDefined } from "../../utils";
import { EnclavesTable } from "../../components/enclaves/EnclavesTable";
import { FiPlus, FiTrash2 } from "react-icons/fi";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const kurtosisClient = useKurtosisClient();
  const [enclaves, setEnclaves] = useState<EnclaveFullInfo[]>();
  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  useEffect(() => {
    (async () => {
      const enclavesResponse = await kurtosisClient.getEnclaves();
      const enclaves = Object.values(enclavesResponse.enclaveInfo);
      const [starlarkRuns, services, filesAndArtifacts] = await Promise.all([
        Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
        Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
        Promise.all(enclaves.map((enclave) => kurtosisClient.listFilesArtifactNamesAndUuids(enclave))),
      ]);

      setEnclaves(
        enclaves.map((enclave, i) => ({
          ...enclave,
          starlarkRun: starlarkRuns[i],
          services: services[i],
          filesAndArtifacts: filesAndArtifacts[i],
        })),
      );
    })();
  }, []);

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
            {!isDefined(enclaves) && (
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
