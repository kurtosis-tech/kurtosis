import { Flex, Spinner, TabPanel, TabPanels, Tabs, Text } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FunctionComponent } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AppPageLayout } from "../../../../components/AppLayout";
import { HoverLineTabList } from "../../../../components/HoverLineTabList";
import { KurtosisAlert } from "../../../../components/KurtosisAlert";
import { isDefined } from "../../../../utils";
import { useFullEnclave } from "../../../EmuiAppContext";
import { EnclaveFullInfo } from "../../types";
import { ServiceLogs } from "./logs/ServiceLogs";
import { ServiceOverview } from "./overview/ServiceOverview";

const tabs: { path: string; element: FunctionComponent<{ enclave: EnclaveFullInfo; service: ServiceInfo }> }[] = [
  { path: "overview", element: ServiceOverview },
  { path: "logs", element: ServiceLogs },
];

export const Service = () => {
  const { enclaveUUID, serviceUUID } = useParams();
  const enclave = useFullEnclave(enclaveUUID || "unknown");

  if (enclave.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={"Enclave could not load"} />
      </AppPageLayout>
    );
  }

  if (!isDefined(enclave.value.services)) {
    return (
      <AppPageLayout>
        <Spinner />
      </AppPageLayout>
    );
  }

  if (enclave.value.services.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={"Services for enclave could not load"} />
      </AppPageLayout>
    );
  }

  const service = Object.values(enclave.value.services.value.serviceInfo).find(
    (service) => service.shortenedUuid === serviceUUID,
  );
  if (!isDefined(service)) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={`Could not find service ${serviceUUID}`} />
      </AppPageLayout>
    );
  }

  return <ServiceImpl enclave={enclave.value} service={service} />;
};

type ServiceImplProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

const ServiceImpl = ({ enclave, service }: ServiceImplProps) => {
  const navigator = useNavigate();
  const params = useParams();
  const activeTab = params.activeTab || "overview";
  const activeIndex = tabs.findIndex((tab) => tab.path === activeTab);

  const handleTabChange = (newTabIndex: number) => {
    const tab = tabs[newTabIndex];
    navigator(`/enclave/${enclave.shortenedUuid}/service/${service.shortenedUuid}/${tab.path}`);
  };

  return (
    <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange}>
      <AppPageLayout>
        <Flex alignItems={"center"} gap={"8px"}>
          <Text as={"span"} fontSize={"lg"} fontWeight={"md"} mb={"4px"}>
            {service.name}
          </Text>
          <HoverLineTabList tabs={tabs.map(({ path }) => path)} activeTab={activeTab} />
        </Flex>
        <TabPanels>
          {tabs.map((tab) => (
            <TabPanel key={tab.path}>
              <tab.element enclave={enclave} service={service} />
            </TabPanel>
          ))}
        </TabPanels>
      </AppPageLayout>
    </Tabs>
  );
};
