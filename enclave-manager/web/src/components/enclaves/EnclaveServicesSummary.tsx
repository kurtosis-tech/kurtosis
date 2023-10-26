import { Button, ButtonGroup, Text, Tooltip } from "@chakra-ui/react";
import { ServiceInfo, ServiceStatus } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "../../utils";

type ServicesSummaryProps = {
  services: ServiceInfo[];
};

export const EnclaveServicesSummary = ({ services }: ServicesSummaryProps) => {
  const runningServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.RUNNING).length;
  const stopppedServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.STOPPED).length;
  const unknownServices = services.filter(({ serviceStatus }) => serviceStatus === ServiceStatus.UNKNOWN).length;

  if (runningServices + stopppedServices + unknownServices === 0) {
    return (
      <Text fontSize={"12px"} as={"i"}>
        No Services
      </Text>
    );
  }

  const tooltipLabel = [
    runningServices > 0 ? `${runningServices} running` : null,
    stopppedServices > 0 ? `${stopppedServices} stopped` : null,
    unknownServices > 0 ? `${unknownServices} unknown` : null,
  ]
    .filter(isDefined)
    .join(", ");

  return (
    <Tooltip label={tooltipLabel} size={"xs"}>
      <ButtonGroup size={"xs"} variant={"solid"}>
        {runningServices > 0 && <Button colorScheme={"green"}>{runningServices}</Button>}
        {stopppedServices > 0 && <Button colorScheme={"red"}>{stopppedServices}</Button>}
        {unknownServices > 0 && <Button colorScheme={"orange"}>{unknownServices}</Button>}
      </ButtonGroup>
    </Tooltip>
  );
};
