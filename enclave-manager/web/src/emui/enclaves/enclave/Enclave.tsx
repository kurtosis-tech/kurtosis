import { Flex, TabPanel, TabPanels, Tabs, Text } from "@chakra-ui/react";
import { useNavigate, useParams } from "react-router-dom";

import { FunctionComponent, useState } from "react";
import { AppPageLayout } from "../../../components/AppLayout";
import { EditEnclaveButton } from "../../../components/enclaves/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../../components/enclaves/widgets/DeleteEnclavesButton";
import { FeatureNotImplementedModal } from "../../../components/FeatureNotImplementedModal";
import { HoverLineTabList } from "../../../components/HoverLineTabList";
import { KurtosisAlert } from "../../../components/KurtosisAlert";
import { useFullEnclave } from "../../EmuiAppContext";
import { EnclaveFullInfo } from "../types";
import { EnclaveOverview } from "./overview/EnclaveOverview";

const tabs: { path: string; element: FunctionComponent<{ enclave: EnclaveFullInfo }> }[] = [
  { path: "overview", element: EnclaveOverview },
];

export const Enclave = () => {
  const { enclaveUUID } = useParams();
  const enclave = useFullEnclave(enclaveUUID || "unknown");

  if (enclave.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={"Enclave could not load"} />
      </AppPageLayout>
    );
  }

  return <EnclaveImpl enclave={enclave.value} />;
};

type EnclaveImplProps = {
  enclave: EnclaveFullInfo;
};

const EnclaveImpl = ({ enclave }: EnclaveImplProps) => {
  const navigator = useNavigate();
  const params = useParams();
  const activeTab = params.activeTab || "overview";
  const activeIndex = tabs.findIndex((tab) => tab.path === activeTab);

  const [unavailableModalState, setUnavailableModalState] = useState<
    { isOpen: false } | { isOpen: true; featureName: string; message?: string; issueUrl: string }
  >({ isOpen: false });

  const handleTabChange = (newTabIndex: number) => {
    const tab = tabs[newTabIndex];
    navigator(`/enclave/${enclave.shortenedUuid}/${tab.path}`);
  };

  return (
    <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange}>
      <AppPageLayout>
        <Flex justifyContent={"space-between"} alignItems={"flex-end"} width={"100%"}>
          <Flex alignItems={"center"} gap={"8px"}>
            <Text as={"span"} fontSize={"lg"} fontWeight={"md"} mb={"4px"}>
              {enclave.name}
            </Text>
            <HoverLineTabList tabs={tabs.map(({ path }) => path)} activeTab={activeTab} />
          </Flex>
          <Flex gap={"8px"} alignItems={"center"} pb={"16px"}>
            <DeleteEnclavesButton enclaves={[enclave]} />
            <EditEnclaveButton enclave={enclave} />
          </Flex>
          <FeatureNotImplementedModal
            featureName={unavailableModalState.isOpen ? unavailableModalState.featureName : ""}
            message={unavailableModalState.isOpen ? unavailableModalState.message : ""}
            isOpen={unavailableModalState.isOpen}
            issueUrl={unavailableModalState.isOpen ? unavailableModalState.issueUrl : ""}
            onClose={() => setUnavailableModalState({ isOpen: false })}
          />
        </Flex>
        <TabPanels>
          {tabs.map((tab) => (
            <TabPanel key={tab.path}>
              <tab.element enclave={enclave} />
            </TabPanel>
          ))}
        </TabPanels>
      </AppPageLayout>
    </Tabs>
  );
};
