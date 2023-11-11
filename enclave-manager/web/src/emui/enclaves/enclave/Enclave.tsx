import { Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { Await, useActionData, useNavigate, useParams, useRouteLoaderData } from "react-router-dom";

import { FunctionComponent, Suspense, useEffect, useState } from "react";
import { EditEnclaveButton } from "../../../components/enclaves/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../../components/enclaves/widgets/DeleteEnclavesButton";
import { FeatureNotImplementedModal } from "../../../components/FeatureNotImplementedModal";
import { KurtosisAlert } from "../../../components/KurtosisAlert";
import { isDefined } from "../../../utils";
import { EnclaveFullInfo } from "../types";
import { RunStarlarkResolvedType } from "./action";
import { EnclaveLoaderResolved } from "./loader";
import { EnclaveLogs } from "./logs/EnclaveLogs";
import { EnclaveOverview } from "./overview/EnclaveOverview";

const tabs: { path: string; element: FunctionComponent<{ enclave: EnclaveFullInfo }> }[] = [
  { path: "overview", element: EnclaveOverview },
  { path: "logs", element: EnclaveLogs },
];

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

type MaybeEnclaveImplProps = {
  enclave: EnclaveLoaderResolved["data"]["enclave"];
};

const MaybeEnclaveImpl = ({ enclave: enclaveResult }: MaybeEnclaveImplProps) => {
  const { enclaveUUID } = useParams();

  if (!isDefined(enclaveResult)) {
    return <KurtosisAlert message={`Could not find enclave ${enclaveUUID}`} />;
  }

  if (enclaveResult.isErr) {
    return <KurtosisAlert message={"Enclave could not load"} />;
  }

  return <EnclaveImpl enclave={enclaveResult.value} />;
};

type EnclaveImplProps = {
  enclave: EnclaveFullInfo;
};

const EnclaveImpl = ({ enclave }: EnclaveImplProps) => {
  const navigator = useNavigate();
  const params = useParams();
  const actionData = useActionData() as undefined | RunStarlarkResolvedType;
  const activeTab = params.activeTab || "overview";
  const activeIndex = tabs.findIndex((tab) => tab.path === activeTab);

  const [unavailableModalState, setUnavailableModalState] = useState<
    { isOpen: false } | { isOpen: true; featureName: string; message?: string; issueUrl: string }
  >({ isOpen: false });

  const handleTabChange = (newTabIndex: number) => {
    const tab = tabs[newTabIndex];
    if (tab.path === "logs" && !isDefined(actionData)) {
      setUnavailableModalState({
        isOpen: true,
        featureName: "Enclave Logs",
        issueUrl: "https://github.com/kurtosis-tech/kurtosis/issues/1721",
        message:
          "Enclave logs are currently only viewable during configuration. Please upvote this feature request if you'd like enclave logs to be persisted.",
      });
      return;
    }
    navigator(`/enclave/${enclave.shortenedUuid}/${tab.path}`);
  };

  useEffect(() => {
    if (isDefined(actionData)) {
      navigator(`/enclave/${enclave.shortenedUuid}/logs`, { state: actionData, replace: true });
    }
  }, [navigator, actionData, activeIndex, enclave.shortenedUuid]);

  return (
    <Flex direction="column" width={"100%"}>
      <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange}>
        <TabList>
          <Flex justifyContent={"space-between"} width={"100%"}>
            <TabList>
              {tabs.map((tab) => (
                <Tab key={tab.path}>{tab.path}</Tab>
              ))}
            </TabList>
            <Flex gap={"8px"} alignItems={"center"}>
              <DeleteEnclavesButton enclaves={[enclave]} />
              <EditEnclaveButton enclave={enclave} />
            </Flex>
          </Flex>
        </TabList>
        <TabPanels>
          {tabs.map((tab) => (
            <TabPanel key={tab.path}>
              <tab.element enclave={enclave} />
            </TabPanel>
          ))}
        </TabPanels>
      </Tabs>
      <FeatureNotImplementedModal
        featureName={unavailableModalState.isOpen ? unavailableModalState.featureName : ""}
        message={unavailableModalState.isOpen ? unavailableModalState.message : ""}
        isOpen={unavailableModalState.isOpen}
        issueUrl={unavailableModalState.isOpen ? unavailableModalState.issueUrl : ""}
        onClose={() => setUnavailableModalState({ isOpen: false })}
      />
    </Flex>
  );
};
