import { Flex, Spinner, Tag, TagProps, Tooltip } from "@chakra-ui/react";
import { ServiceInfo, ServiceStatus } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "kurtosis-ui-components";

type ServicesSummaryProps = {
  services: "loading" | ServiceInfo[] | null;
};

export const EnclaveServicesSummary = ({ services }: ServicesSummaryProps) => {
  if (!isDefined(services)) {
    return <Tag>Unknown</Tag>;
  }

  if (services === "loading") {
    return <Spinner size={"xs"} />;
  }

  const runningServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.RUNNING).length;
  const stopppedServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.STOPPED).length;
  const unknownServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.UNKNOWN).length;

  const totalServices = runningServices + stopppedServices + unknownServices;

  const tooltipLabel = [
    runningServices > 0 ? `${runningServices} running` : null,
    stopppedServices > 0 ? `${stopppedServices} stopped` : null,
    unknownServices > 0 ? `${unknownServices} unknown` : null,
  ]
    .filter(isDefined)
    .join(", ");

  const tagProps: Partial<TagProps> = {
    variant: "solid",
    fontSize: "xs",
    fontWeight: "semibold",
  };

  return (
    <Tooltip label={tooltipLabel} size={"xs"}>
      <Flex justifyContent={"center"}>
        {totalServices === 0 && (
          <Tag color={"#A0AEC0"} {...tagProps}>
            NONE
          </Tag>
        )}
        {runningServices > 0 && (
          <Tag colorScheme={"green"} {...tagProps}>
            {runningServices}
          </Tag>
        )}
        {stopppedServices > 0 && (
          <Tag colorScheme={"red"} {...tagProps}>
            {stopppedServices}
          </Tag>
        )}
        {unknownServices > 0 && (
          <Tag colorScheme={"orange"} {...tagProps}>
            {unknownServices}
          </Tag>
        )}
      </Flex>
    </Tooltip>
  );
};
