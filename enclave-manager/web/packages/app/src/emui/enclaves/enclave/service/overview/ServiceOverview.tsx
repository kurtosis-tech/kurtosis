import { Flex, Grid, GridItem, Icon, Text } from "@chakra-ui/react";
import { Container, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FileDisplay, isDefined, KurtosisAlert, TitledBox, ValueCard } from "kurtosis-ui-components";
import { useMemo } from "react";
import { IoLogoDocker } from "react-icons/io5";
import { PortsTable } from "../../../components/tables/PortsTable";
import { ServiceStatusTag } from "../../../components/widgets/ServiceStatus";
import { EnclaveFullInfo } from "../../../types";

type ServiceOverviewProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

export const ServiceOverview = ({ service, enclave }: ServiceOverviewProps) => {
  return (
    <Flex flexDirection={"column"} gap={"32px"}>
      <Grid templateColumns={"repeat(4, 1fr)"} gap={"32px"}>
        <GridItem>
          <ValueCard title={"Name"} value={service.name} copyEnabled />
        </GridItem>
        <GridItem>
          <ValueCard title={"UUID"} value={service.shortenedUuid} copyEnabled />
        </GridItem>
        <GridItem>
          <ValueCard title={"Status"} value={<ServiceStatusTag status={service.serviceStatus} variant={"asText"} />} />
        </GridItem>
        <GridItem>
          <ValueCard
            title={"Image"}
            value={
              <Flex alignItems={"center"} gap={"8px"}>
                <Icon as={IoLogoDocker} />
                <Text>{service.container?.imageName || "unknown"}</Text>
              </Flex>
            }
          />
        </GridItem>
      </Grid>
      <TitledBox title={"Ports"}>
        <PortsTable
          enclaveUUID={enclave.enclaveUuid}
          serviceUUID={service.serviceUuid}
          privatePorts={service.privatePorts}
          publicPorts={service.maybePublicPorts}
          publicIp={service.maybePublicIpAddr}
        />
      </TitledBox>
      {isDefined(service.container) && (
        <ContainerOverview serviceName={service.name} enclaveName={enclave.name} container={service.container} />
      )}
      {!isDefined(service.container) && (
        <KurtosisAlert message={"No container details are available for this service."} />
      )}
    </Flex>
  );
};

type ContainerOverviewProps = {
  enclaveName: string;
  serviceName: string;
  container: Container;
};

const ContainerOverview = ({ enclaveName, container, serviceName }: ContainerOverviewProps) => {
  const environmentJson = useMemo(() => JSON.stringify(container.envVars, undefined, 4), [container]);
  const cmdJson = useMemo(() => JSON.stringify(container.cmdArgs, undefined, 4), [container]);
  const entrypointJson = useMemo(() => JSON.stringify(container.entrypointArgs, undefined, 4), [container]);

  return (
    <TitledBox title={"Detailed Info"}>
      <Grid gridColumnGap={"32px"} gridTemplateColumns={"1fr 1fr"} width={"100%"}>
        <GridItem display={"flex"} flexDirection={"column"} gap={"16px"} height={"100%"}>
          <FileDisplay
            value={entrypointJson}
            title={"ENTRYPOINT"}
            filename={`${enclaveName}--${serviceName}-entrypoint.json`}
          />
          <FileDisplay value={cmdJson} title={"CMD"} filename={`${enclaveName}--${serviceName}-cmd.json`} />
        </GridItem>
        <GridItem>
          <FileDisplay
            value={environmentJson}
            title={"ENVIRONMENT"}
            filename={`${enclaveName}--${serviceName}-env.json`}
          />
        </GridItem>
      </Grid>
    </TitledBox>
  );
};
