import { Card, Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { Suspense, useEffect, useState } from "react";
import { GetEnclavesResponse } from "enclave-manager-sdk/build/engine_service_pb";
import { isDefined } from "../../utils";
import { EnclavesTable } from "../../components/enclaves/EnclavesTable";

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
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosis"}>
        <TabList>
          <Tab>Overview</Tab>
        </TabList>
        <TabPanels>
          <TabPanel>
            {isDefined(encalvesResponse) && (
              <Card w={"100%"}>
                <EnclavesTable enclavesData={encalvesResponse.enclaveInfo} />
              </Card>
            )}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
