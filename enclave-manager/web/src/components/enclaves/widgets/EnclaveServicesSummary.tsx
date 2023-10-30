import { Button, ButtonGroup, Tag, Tooltip } from "@chakra-ui/react";
import { ServiceInfo, ServiceStatus } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "../../../utils";

type ServicesSummaryProps = {
  services: ServiceInfo[] | null;
};

export const EnclaveServicesSummary = ({ services }: ServicesSummaryProps) => {
  if (!isDefined(services)) {
    return <Tag>Unknown</Tag>;
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

  return (
    <Tooltip label={tooltipLabel} size={"xs"}>
      <ButtonGroup size={"xs"} isAttached variant={"solid"}>
        {totalServices === 0 && <Button color={"#A0AEC0"}>NONE</Button>}
        {runningServices > 0 && <Button colorScheme={"green"}>{runningServices}</Button>}
        {stopppedServices > 0 && <Button colorScheme={"red"}>{stopppedServices}</Button>}
        {unknownServices > 0 && <Button colorScheme={"orange"}>{unknownServices}</Button>}
      </ButtonGroup>
    </Tooltip>
  );
};
