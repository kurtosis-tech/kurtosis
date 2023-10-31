import { Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { Await, useActionData, useParams, useRouteLoaderData } from "react-router-dom";
import { EnclaveOverview } from "../../../components/enclaves/EnclaveOverview";

import { Suspense, useEffect } from "react";
import { EditEnclaveButton } from "../../../components/enclaves/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../../components/KurtosisAlert";
import { isDefined } from "../../../utils";
import { EnclaveFullInfo } from "../types";
import { EnclaveActionResolvedType } from "./action";
import { EnclaveLoaderResolved } from "./loader";

export const Enclave = () => {
  const { data } = useRouteLoaderData("enclave") as EnclaveLoaderResolved;

  return (
    <Suspense
      fallback={
        <Flex justifyContent={"center"} p={"20px"}>
          <Spinner size={"xl"} />
        </Flex>
      }
    >
      <Await resolve={data} children={(data) => <MaybeEnclaveImpl enclave={data.enclave} />} />
    </Suspense>
  );
};

type EnclaveImplProps = {
  enclave: EnclaveLoaderResolved["data"]["enclave"];
};

const MaybeEnclaveImpl = ({ enclave: enclaveResult }: EnclaveImplProps) => {
  const { enclaveUUID, activeTab } = useParams();

  if (!isDefined(enclaveResult)) {
    return <KurtosisAlert message={`Could not find enclave ${enclaveUUID}`} />;
  }

  if (enclaveResult.isErr) {
    return <KurtosisAlert message={"Enclave could not load"} />;
  }

  return <EnclaveImpl enclave={enclaveResult.value} />;
};

type EnclaveImpl = {
  enclave: EnclaveFullInfo;
};

const EnclaveImpl = ({ enclave }: EnclaveImpl) => {
  const actionData = useActionData() as undefined | EnclaveActionResolvedType;

  useEffect(() => {
    if (actionData) {
      (async () => {
        for await (const line of actionData.logs) {
          console.log(line.runResponseLine.value);
        }
      })();
    }
  }, [actionData]);

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
              <EditEnclaveButton enclave={enclave} />
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
