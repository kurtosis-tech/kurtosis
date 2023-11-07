import { ButtonGroup, Card, Flex, Grid, GridItem, Icon, Text } from "@chakra-ui/react";
import { Container, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { IoLogoDocker } from "react-icons/io5";
import { ServiceStatusTag } from "../../../../../components/enclaves/widgets/ServiceStatus";
import { FLEX_STANDARD_GAP } from "../../../../../components/theme/constants";
import { ValueCard } from "../../../../../components/ValueCard";
import { isDefined } from "../../../../../utils";
import { KurtosisAlert } from "../../../../../components/KurtosisAlert";
import { CopyButton } from "../../../../../components/CopyButton";
import { DownloadButton } from "../../../../../components/DownloadButton";
import { useMemo } from "react";
import { CodeEditor } from "../../../../../components/CodeEditor";
import { FileDisplay } from "../../../../../components/FileDisplay";
import { PortsTable } from "../../../../../components/enclaves/tables/PortsTable";
import { TitledCard } from "../../../../../components/TitledCard";

type ServiceOverviewProps = {
  service: ServiceInfo;
};

export const ServiceOverview = ({ service }: ServiceOverviewProps) => {
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
      <TitledCard title={"Public Ports"}>
        <PortsTable ports={Object.values(service.maybePublicPorts)} ip={service.maybePublicIpAddr} isPublic />
      </TitledCard>
      <TitledCard title={"Private Ports"}>
        <PortsTable ports={Object.values(service.privatePorts)} ip={service.privateIpAddr} />
      </TitledCard>
      {isDefined(service.container) && <ContainerOverview container={service.container} />}
      {!isDefined(service.container) && (
        <KurtosisAlert message={"No container details are available for this service."} />
      )}
    </Flex>
  );
};

type ContainerOverviewProps = {
  container: Container;
};

const ContainerOverview = ({ container }: ContainerOverviewProps) => {
  const environmentJson = useMemo(() => JSON.stringify(container.envVars, undefined, 4), [container]);
  const cmdJson = useMemo(() => JSON.stringify(container.cmdArgs, undefined, 4), [container]);
  const entrypointJson = useMemo(() => JSON.stringify(container.entrypointArgs, undefined, 4), [container]);

  const filePrefix = container.imageName.replaceAll(/:/g, "_");

  return (
    <Flex flexDirection={"column"} gap={"32px"}>
      <Text fontSize={"md"} fontWeight={"semibold"}>
        Detailed Info
      </Text>
      <Grid gridColumnGap={"32px"} gridTemplateColumns={"1fr 1fr"}>
        <GridItem display={"flex"} flexDirection={"column"} gap={"16px"}>
          <FileDisplay value={entrypointJson} title={"Entrypoint"} filename={`${filePrefix}-entrypoint.json`} />
          <FileDisplay value={cmdJson} title={"CMD"} filename={`${filePrefix}-cmd.json`} />
        </GridItem>
        <GridItem>
          <FileDisplay value={environmentJson} title={"Environment"} filename={`${filePrefix}-env.json`} />
        </GridItem>
      </Grid>
    </Flex>
  );
};
