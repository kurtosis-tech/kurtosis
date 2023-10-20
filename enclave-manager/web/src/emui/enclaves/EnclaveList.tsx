import { Card, Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useKurtosisClient } from "../../client/KurtosisClientContext";
import { useEffect } from "react";

export const EnclaveList = () => {
  const kurtosisClient = useKurtosisClient();

  useEffect(() => {
    (async () => {
      const enclaves = await kurtosisClient.getEnclaves();
      console.log(enclaves);
    })();
  }, []);

  return (
    <Flex direction="column">
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosis"}>
        <TabList>
          <Tab>Overview</Tab>
        </TabList>
        <TabPanels>
          <Card w={"100%"}>Hello table</Card>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
