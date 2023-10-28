import { Button, ButtonGroup, Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { Suspense, useEffect, useState } from "react";
import { FiPlus } from "react-icons/fi";
import { Await, useRouteLoaderData } from "react-router-dom";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { EnclavesLoaderResolved } from "./loader";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const { enclaves } = useRouteLoaderData("enclaves") as EnclavesLoaderResolved;

  return (
    <Suspense
      fallback={
        <Flex justifyContent={"center"} p={"20px"}>
          <Spinner size={"xl"} />
        </Flex>
      }
    >
      <Await resolve={enclaves} children={(enclaves) => <EnclaveListImpl enclaves={enclaves} />} />
    </Suspense>
  );
};

type EnclaveListImplProps = {
  enclaves: EnclavesLoaderResolved["enclaves"];
};

const EnclaveListImpl = ({ enclaves }: EnclaveListImplProps) => {
  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  useEffect(() => {
    setSelectedEnclaves([]);
  }, [enclaves]);

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
            <Button colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
              Create Enclave
            </Button>
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
