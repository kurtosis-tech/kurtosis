import { Box, Button, Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { FiEdit2, FiTrash2 } from "react-icons/fi";
import { LoaderFunctionArgs, useParams, useRouteLoaderData } from "react-router-dom";
import { getKurtosisClient } from "../../client/KurtosisClientContext";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { isDefined } from "../../utils";
import { enclavesLoader } from "./EnclaveList";

export const enclaveLoader = async ({ params }: LoaderFunctionArgs): Promise<{ routeName: string }> => {
  const uuid = params.uuid;

  if (!isDefined(uuid)) {
    return {
      routeName: "Missing uuid",
    };
  }

  const kurtosisClient = await getKurtosisClient();
  const enclavesResult = await kurtosisClient.getEnclaves();
  if (enclavesResult.isErr) {
    return {
      routeName: uuid,
    };
  }

  const enclave = Object.values(enclavesResult.value.enclaveInfo).find((enclave) => enclave.enclaveUuid === uuid);
  if (!isDefined(enclave)) {
    return {
      routeName: uuid,
    };
  }

  return {
    routeName: enclave.name,
  };
};

export const enclaveTabLoader = async ({ params }: LoaderFunctionArgs): Promise<{ routeName: string }> => {
  const activeTab = params.activeTab;

  switch (activeTab?.toLowerCase()) {
    case "overview":
      return { routeName: "Overview" };
    case "source":
      return { routeName: "Source" };
    default:
      return { routeName: "Overview" };
  }
};

export const Enclave = () => {
  const { uuid, activeTab } = useParams();
  const enclaves = useRouteLoaderData("enclaves") as Awaited<ReturnType<typeof enclavesLoader>>;
  if (enclaves.isErr) {
    return <KurtosisAlert message={"Enclaves could not load"} />;
  }
  const enclave = enclaves.value.find((e) => e.enclaveUuid === uuid);
  if (!isDefined(enclave)) {
    return <KurtosisAlert message={`Could not find enclave ${uuid}`} />;
  }
  return (
    <Flex direction="column" width={"100%"}>
      <Tabs>
        <TabList>
          <Flex justifyContent={"space-between"} width={"100%"}>
            <TabList>
              <Tab>Overview</Tab>
              <Tab>Source</Tab>
            </TabList>
            <Flex gap={"8px"} alignItems={"center"}>
              <Button colorScheme={"red"} leftIcon={<FiTrash2 />} variant={"kurtosisOutline"} size={"md"}>
                Delete
              </Button>
              <Button colorScheme={"blue"} leftIcon={<FiEdit2 />} variant={"kurtosisOutline"} size={"md"}>
                Edit
              </Button>
            </Flex>
          </Flex>
        </TabList>
        <TabPanels>
          <TabPanel>
            <Box>Enclave {enclave.name}</Box>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
