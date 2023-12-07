import { Tag, Tooltip } from "@chakra-ui/react";
import { ServiceStatus } from "enclave-manager-sdk/build/api_container_service_pb";

export function serviceStatusToString(status: ServiceStatus) {
  switch (status) {
    case ServiceStatus.RUNNING:
      return "Running";
    case ServiceStatus.STOPPED:
      return "Stopped";
    case ServiceStatus.UNKNOWN:
      return "Unknown";
  }
}

export function serviceStatusToColorScheme(status: ServiceStatus) {
  switch (status) {
    case ServiceStatus.RUNNING:
      return "green";
    case ServiceStatus.STOPPED:
      return "red";
    case ServiceStatus.UNKNOWN:
      return "orange";
  }
}

type ServiceStatusTagProps = {
  status: ServiceStatus;
  variant?: string;
};

export const ServiceStatusTag = ({ status, variant }: ServiceStatusTagProps) => {
  const display = serviceStatusToString(status);
  const colorScheme = serviceStatusToColorScheme(status);

  return (
    <Tooltip label={"The status of the container providing this service."} openDelay={1000}>
      <Tag variant={variant} colorScheme={colorScheme}>
        {display}
      </Tag>
    </Tooltip>
  );
};
