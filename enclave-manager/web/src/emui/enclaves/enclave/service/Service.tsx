import { Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FunctionComponent } from "react";
import { useNavigate, useParams } from "react-router-dom";
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
    return <KurtosisAlert message={"Enclave could not load"} />;
  }

  if (!isDefined(enclave.value.services)) {
    return <Spinner />;
  }

  if (enclave.value.services.isErr) {
    return <KurtosisAlert message={"Services for enclave could not load"} />;
  }

  const service = Object.values(enclave.value.services.value.serviceInfo).find(
    (service) => service.shortenedUuid === serviceUUID,
  );
  if (!isDefined(service)) {
    return <KurtosisAlert message={`Could not find service ${serviceUUID}`} />;
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
    <Flex direction="column" width={"100%"}>
      <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange}>
        <TabList>
          <TabList>
            {tabs.map((tab) => (
              <Tab key={tab.path}>{tab.path}</Tab>
            ))}
          </TabList>
        </TabList>
        <TabPanels>
          {tabs.map((tab) => (
            <TabPanel key={tab.path}>
              <tab.element enclave={enclave} service={service} />
            </TabPanel>
          ))}
        </TabPanels>
      </Tabs>
    </Flex>
  );
};
