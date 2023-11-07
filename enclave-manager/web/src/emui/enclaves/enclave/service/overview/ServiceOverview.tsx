import { Flex, Grid, GridItem, Icon, Text } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { IoLogoDocker } from "react-icons/io5";
import { ServiceStatusTag } from "../../../../../components/enclaves/widgets/ServiceStatus";
import { FLEX_STANDARD_GAP } from "../../../../../components/theme/constants";
import { ValueCard } from "../../../../../components/ValueCard";

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
              <Flex>
                <Icon as={IoLogoDocker} />
                <Text>{service.container?.imageName || "unknown"}</Text>
              </Flex>
            }
          />
        </GridItem>
      </Grid>
    </Flex>
  );
};
