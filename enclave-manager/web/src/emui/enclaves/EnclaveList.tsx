import { Button, ButtonGroup, Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useState } from "react";
import { CreateEnclaveButton } from "../../components/enclaves/CreateEnclaveButton";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { useFullEnclaves } from "../EmuiAppContext";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const enclaves = useFullEnclaves();

  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  return (
    <Flex direction="column">
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosisGreen"}>
        <Flex justifyContent={"space-between"}>
          <TabList>
            <Tab>Enclaves</Tab>
          </TabList>
          <Flex gap={"24px"} alignItems={"center"}>
            {selectedEnclaves.length > 0 && (
              <ButtonGroup isAttached variant={"kurtosisGroupOutline"} size={"sm"}>
                <Button variant={"kurtosisDisabled"} colorScheme={"gray"}>
                  {selectedEnclaves.length} selected
                </Button>
                <DeleteEnclavesButton enclaves={selectedEnclaves} />
              </ButtonGroup>
            )}
            <CreateEnclaveButton />
          </Flex>
        </Flex>
        <TabPanels>
          <TabPanel>
            {enclaves.isOk && (
              <EnclavesTable
                enclavesData={enclaves.value}
                selection={selectedEnclaves}
                onSelectionChange={setSelectedEnclaves}
              />
            )}
            {enclaves.isErr && <KurtosisAlert message={enclaves.error} />}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
