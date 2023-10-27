import { Button, Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { FiEdit2 } from "react-icons/fi";
import { Await, LoaderFunctionArgs, useParams, useRouteLoaderData } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { EnclaveOverview } from "../../components/enclaves/EnclaveOverview";

import { Suspense } from "react";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { isDefined } from "../../utils";
import { EnclavesLoaderResolved } from "./EnclaveList";

export const enclaveLoader =
  (kurtosisClient: KurtosisClient) =>
  async ({ params }: LoaderFunctionArgs): Promise<{ routeName: string }> => {
    const uuid = params.enclaveUUID;

    if (!isDefined(uuid)) {
      return {
        routeName: "Missing uuid",
      };
    }

    const enclavesResult = await kurtosisClient.getEnclaves();
    if (enclavesResult.isErr) {
      return {
        routeName: uuid,
      };
    }

    const enclave = Object.values(enclavesResult.value.enclaveInfo).find((enclave) => enclave.shortenedUuid === uuid);
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
  const { enclaves } = useRouteLoaderData("enclaves") as EnclavesLoaderResolved;

  return (
    <Suspense
      fallback={
        <Flex justifyContent={"center"} p={"20px"}>
          <Spinner size={"xl"} />
        </Flex>
      }
    >
      <Await resolve={enclaves} children={(enclaves) => <EnclaveImpl enclaves={enclaves} />} />
    </Suspense>
  );
};

type EnclaveImplProps = {
  enclaves: EnclavesLoaderResolved["enclaves"];
};

const EnclaveImpl = ({ enclaves }: EnclaveImplProps) => {
  const { enclaveUUID, activeTab } = useParams();

  if (enclaves.isErr) {
    return <KurtosisAlert message={"Enclaves could not load"} />;
  }
  const enclave = enclaves.value.find((e) => e.shortenedUuid === enclaveUUID);
  if (!isDefined(enclave)) {
    return <KurtosisAlert message={`Could not find enclave ${enclaveUUID}`} />;
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
              <DeleteEnclavesButton enclaves={[enclave]} />
              <Button colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"md"}>
                Edit
              </Button>
            </Flex>
          </Flex>
        </TabList>
        <TabPanels>
          <TabPanel>
            <EnclaveOverview enclave={enclave} />
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
