import { Button, ButtonGroup, Card, Flex, Icon, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { useEffect, useState } from "react";
import { isDefined } from "../../utils";
import { EnclavesTable } from "../../components/enclaves/EnclavesTable";
import { FiPlus, FiTrash2 } from "react-icons/fi";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const kurtosisClient = useKurtosisClient();
  const [enclaves, setEnclaves] = useState<EnclaveFullInfo[]>();

  useEffect(() => {
    (async () => {
      const enclavesResponse = await kurtosisClient.getEnclaves();
      const enclaves = Object.values(enclavesResponse.enclaveInfo);
      const [starlarkRuns, services] = await Promise.all([
        Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
        Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
      ]);

      setEnclaves(enclaves.map((enclave, i) => ({ ...enclave, starlarkRun: starlarkRuns[i], services: services[i] })));
    })();
  }, []);

  return (
    <Flex direction="column">
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosisGreen"}>
        <Flex justifyContent={"space-between"}>
          <TabList>
            <Tab>Enclaves</Tab>
            <Tab>Overview</Tab>
          </TabList>
          <Flex gap={"24px"} alignItems={"center"}>
            <ButtonGroup isAttached variant={"kurtosisGroupOutline"} size={"sm"}>
              <Button colorScheme={"blue"}>3 selected</Button>
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
          <TabPanel>{isDefined(enclaves) && <EnclavesTable enclavesData={enclaves} />}</TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
