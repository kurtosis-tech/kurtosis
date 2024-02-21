import { Flex, Spinner, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { AppPageLayout, HoverLineTabList, isDefined, KurtosisAlert, PageTitle } from "kurtosis-ui-components";
import { FunctionComponent } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { EditEnclaveButton } from "../../components/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../components/widgets/DeleteEnclavesButton";
import { EnclaveFullInfo } from "../../types";
import { useEnclaveFromParams } from "../EnclaveRouteContext";
import { ServiceLogs } from "./logs/ServiceLogs";
import { ServiceOverview } from "./overview/ServiceOverview";

const tabs: {
  path: string;
  element: FunctionComponent<{ enclave: EnclaveFullInfo; service: ServiceInfo }>;
}[] = [
  { path: "overview", element: ServiceOverview },
  { path: "logs", element: ServiceLogs },
];

export const Service = () => {
  const { serviceUUID } = useParams();
  const enclave = useEnclaveFromParams();

  if (!isDefined(enclave.services)) {
    return (
      <AppPageLayout>
        <Spinner />
      </AppPageLayout>
    );
  }

  if (enclave.services.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={"Services for enclave could not load"} />
      </AppPageLayout>
    );
  }

  const service = Object.values(enclave.services.value.serviceInfo).find(
    (service) => service.shortenedUuid === serviceUUID,
  );
  if (!isDefined(service)) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={`Could not find service ${serviceUUID}`} />
      </AppPageLayout>
    );
  }

  return <ServiceImpl enclave={enclave} service={service} />;
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
    <Tabs isManual isLazy index={activeIndex} onChange={handleTabChange} variant={"kurtosisHeaderLine"}>
      <AppPageLayout>
        <Flex justifyContent={"space-between"} alignItems={"flex-end"} width={"100%"}>
          <Flex alignItems={"center"} gap={"8px"}>
            <PageTitle>{service.name}</PageTitle>
            <HoverLineTabList tabs={tabs.map(({ path }) => path)} activeTab={activeTab} />
          </Flex>
          <Flex gap={"8px"} alignItems={"center"} pb={"16px"}>
            <DeleteEnclavesButton enclaves={[enclave]} />
            <EditEnclaveButton enclave={enclave} />
          </Flex>
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
