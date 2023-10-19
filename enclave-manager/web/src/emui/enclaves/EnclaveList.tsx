import { Card, Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";

export const EnclaveList = () => {
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
