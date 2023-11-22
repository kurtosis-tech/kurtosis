import { Flex, Grid, GridItem, Icon, Text } from "@chakra-ui/react";
import { Container, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { IoLogoDocker } from "react-icons/io5";
import { PortsTable } from "../../../../../components/enclaves/tables/PortsTable";
import { ServiceStatusTag } from "../../../../../components/enclaves/widgets/ServiceStatus";
import { FileDisplay } from "../../../../../components/FileDisplay";
import { KurtosisAlert } from "../../../../../components/KurtosisAlert";
import { FLEX_STANDARD_GAP } from "../../../../../components/theme/constants";
import { TitledCard } from "../../../../../components/TitledCard";
import { ValueCard } from "../../../../../components/ValueCard";
import { isDefined } from "../../../../../utils";
import { EnclaveFullInfo } from "../../../types";

type ServiceOverviewProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

export const ServiceOverview = ({ service, enclave }: ServiceOverviewProps) => {
  return (
    <Flex flexDirection={"column"} gap={FLEX_STANDARD_GAP}>
      <Grid templateColumns={"repeat(4, 1fr)"} gap={FLEX_STANDARD_GAP}>
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
      <TitledCard title={"Ports"}>
        <PortsTable
          privatePorts={service.privatePorts}
          publicPorts={service.maybePublicPorts}
          publicIp={service.maybePublicIpAddr}
        />
      </TitledCard>
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
    <Flex flexDirection={"column"} gap={"32px"}>
      <Text fontSize={"md"} fontWeight={"semibold"}>
        Detailed Info
      </Text>
      <Grid gridColumnGap={"32px"} gridTemplateColumns={"1fr 1fr"}>
        <GridItem display={"flex"} flexDirection={"column"} gap={"16px"}>
          <FileDisplay
            value={entrypointJson}
            title={"Entrypoint"}
            filename={`${enclaveName}--${serviceName}-entrypoint.json`}
          />
          <FileDisplay value={cmdJson} title={"CMD"} filename={`${enclaveName}--${serviceName}-cmd.json`} />
        </GridItem>
        <GridItem>
          <FileDisplay
            value={environmentJson}
            title={"Environment"}
            filename={`${enclaveName}--${serviceName}-env.json`}
          />
        </GridItem>
      </Grid>
    </Flex>
  );
};
