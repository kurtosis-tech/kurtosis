import { Button, ButtonGroup, Card, Flex, Icon, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { Suspense, useEffect, useState } from "react";
import { GetEnclavesResponse } from "enclave-manager-sdk/build/engine_service_pb";
import { isDefined } from "../../utils";
import { EnclavesTable } from "../../components/enclaves/EnclavesTable";
import { FiPlus, FiTrash2 } from "react-icons/fi";

export const EnclaveList = () => {
  const kurtosisClient = useKurtosisClient();

  const [encalvesResponse, setEnclavesResponse] = useState<GetEnclavesResponse>();

  const enclaves = kurtosisClient.getEnclaves();

  useEffect(() => {
    (async () => {
      const enclaves = await kurtosisClient.getEnclaves();
      kurtosisClient.getStarlarkRun(Object.values(enclaves.enclaveInfo)[0]!).then(console.log);
      kurtosisClient.getServices(Object.values(enclaves.enclaveInfo)[0]!).then(console.log);
      setEnclavesResponse(enclaves);
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
          <TabPanel>
            {isDefined(encalvesResponse) && <EnclavesTable enclavesData={encalvesResponse.enclaveInfo} />}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
