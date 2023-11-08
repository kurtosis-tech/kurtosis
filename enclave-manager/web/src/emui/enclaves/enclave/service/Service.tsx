import { Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FunctionComponent, Suspense } from "react";
import { Await, useNavigate, useParams, useRouteLoaderData } from "react-router-dom";
import { KurtosisAlert } from "../../../../components/KurtosisAlert";
import { isDefined } from "../../../../utils";
import { EnclaveFullInfo } from "../../types";
import { EnclaveLoaderResolved } from "../loader";
import { ServiceLogs } from "./logs/ServiceLogs";
import { ServiceOverview } from "./overview/ServiceOverview";

const tabs: { path: string; element: FunctionComponent<{ enclave: EnclaveFullInfo; service: ServiceInfo }> }[] = [
  { path: "overview", element: ServiceOverview },
  { path: "logs", element: ServiceLogs },
];

export const Service = () => {
  const { data } = useRouteLoaderData("enclave") as EnclaveLoaderResolved;

  return (
    <Suspense
      fallback={
        <Flex justifyContent={"center"} p={"20px"}>
          <Spinner size={"xl"} />
        </Flex>
      }
    >
      <Await resolve={data} children={(data) => <MaybeServiceImpl enclave={data.enclave} />} />
    </Suspense>
  );
};

type MaybeServiceImplProps = {
  enclave: EnclaveLoaderResolved["data"]["enclave"];
};

const MaybeServiceImpl = ({ enclave: enclaveResult }: MaybeServiceImplProps) => {
  const { enclaveUUID, serviceUUID } = useParams();

  if (!isDefined(enclaveResult)) {
    return <KurtosisAlert message={`Could not find enclave ${enclaveUUID}`} />;
  }

  if (enclaveResult.isErr) {
    return <KurtosisAlert message={"Enclave could not load"} />;
  }

  if (enclaveResult.value.services.isErr) {
    return <KurtosisAlert message={"Services for enclave could not load"} />;
  }

  const service = Object.values(enclaveResult.value.services.value.serviceInfo).find(
    (service) => service.shortenedUuid === serviceUUID,
  );
  if (!isDefined(service)) {
    return <KurtosisAlert message={`Could not find service ${serviceUUID}`} />;
  }

  return <ServiceImpl enclave={enclaveResult.value} service={service} />;
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
