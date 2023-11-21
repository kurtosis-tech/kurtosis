import { Flex, TabPanel, TabPanels, Tabs, Text } from "@chakra-ui/react";
import { Location, useLocation, useNavigate, useParams } from "react-router-dom";

import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { FunctionComponent, useEffect, useState } from "react";
import { EditEnclaveButton } from "../../../components/enclaves/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../../components/enclaves/widgets/DeleteEnclavesButton";
import { FeatureNotImplementedModal } from "../../../components/FeatureNotImplementedModal";
import { HoverLineTabList } from "../../../components/HoverLineTabList";
import { KurtosisAlert } from "../../../components/KurtosisAlert";
import {
  MAIN_APP_LEFT_PADDING,
  MAIN_APP_MAX_WIDTH,
  MAIN_APP_MAX_WIDTH_WITHOUT_PADDING,
  MAIN_APP_RIGHT_PADDING,
  MAIN_APP_TABPANEL_PADDING,
} from "../../../components/theme/constants";
import { isDefined } from "../../../utils";
import { useFullEnclave } from "../../EmuiAppContext";
import { EnclaveFullInfo } from "../types";
import { EnclaveLogs } from "./logs/EnclaveLogs";
import { EnclaveOverview } from "./overview/EnclaveOverview";

const tabs: { path: string; element: FunctionComponent<{ enclave: EnclaveFullInfo }> }[] = [
  { path: "overview", element: EnclaveOverview },
  { path: "logs", element: EnclaveLogs },
];

export const Enclave = () => {
  const { enclaveUUID } = useParams();
  const enclave = useFullEnclave(enclaveUUID || "unknown");

  if (enclave.isErr) {
    return <KurtosisAlert message={"Enclave could not load"} />;
  }

  return <EnclaveImpl enclave={enclave.value} />;
};

type EnclaveImplProps = {
  enclave: EnclaveFullInfo;
};

const EnclaveImpl = ({ enclave }: EnclaveImplProps) => {
  const navigator = useNavigate();
  const params = useParams();
  const location = useLocation() as Location<{ logs: AsyncIterable<StarlarkRunResponseLine> }>;
  const activeTab = params.activeTab || "overview";
  const activeIndex = tabs.findIndex((tab) => tab.path === activeTab);

  const [unavailableModalState, setUnavailableModalState] = useState<
    { isOpen: false } | { isOpen: true; featureName: string; message?: string; issueUrl: string }
  >({ isOpen: false });

  const handleTabChange = (newTabIndex: number) => {
    const tab = tabs[newTabIndex];
    if (tab.path === "logs" && !isDefined(location.state?.logs)) {
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
    if (isDefined(location.state?.logs)) {
      navigator(`/enclave/${enclave.shortenedUuid}/logs`, { state: location.state, replace: true });
    }
  }, [navigator, location.state, activeIndex, enclave.shortenedUuid]);

  return (
    <Flex direction="column" gap={"24px"} width={"100%"} h={"100%"}>
      <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange}>
        <Flex width={"100%"} bg={"gray.850"} pl={MAIN_APP_LEFT_PADDING} pr={MAIN_APP_RIGHT_PADDING}>
          <Flex
            justifyContent={"space-between"}
            alignItems={"flex-end"}
            width={"100%"}
            maxWidth={MAIN_APP_MAX_WIDTH_WITHOUT_PADDING}
          >
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
          </Flex>
        </Flex>
        <TabPanels maxWidth={MAIN_APP_MAX_WIDTH} p={MAIN_APP_TABPANEL_PADDING} w={"100%"} h={"100%"}>
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
