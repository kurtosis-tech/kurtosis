import { Flex, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import Cookies from "js-cookie";
import { useNavigate, useParams } from "react-router-dom";

import { AppPageLayout, HoverLineTabList, KurtosisAlert, PageTitle } from "kurtosis-ui-components";
import { FunctionComponent } from "react";
import { EditEnclaveButton } from "../components/EditEnclaveButton";
import { ConnectEnclaveButton } from "../components/widgets/ConnectEnclaveButton";
import { DeleteEnclavesButton } from "../components/widgets/DeleteEnclavesButton";
import { useFullEnclave } from "../EnclavesContext";
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

  const handleTabChange = (newTabIndex: number) => {
    const tab = tabs[newTabIndex];
    navigator(`/enclave/${enclave.shortenedUuid}/${tab.path}`);
  };

  const instanceUUID = Cookies.get("_kurtosis_instance_id") || "";

  return (
    <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange} variant={"kurtosisHeaderLine"}>
      <AppPageLayout preventPageScroll={activeTab === "logs"}>
        <Flex justifyContent={"space-between"} alignItems={"flex-end"} width={"100%"}>
          <Flex alignItems={"center"} gap={"8px"}>
            <PageTitle>{enclave.name}</PageTitle>
            <HoverLineTabList tabs={tabs.map(({ path }) => path)} activeTab={activeTab} />
          </Flex>
          <Flex gap={"8px"} alignItems={"center"} pb={"16px"}>
            <DeleteEnclavesButton enclaves={[enclave]} />
            <EditEnclaveButton enclave={enclave} />
            <ConnectEnclaveButton enclave={enclave} instanceUUID={instanceUUID} />
          </Flex>
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
