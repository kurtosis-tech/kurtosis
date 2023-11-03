import { Tag } from "@chakra-ui/react";
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

type ServiceStatusTagProps = {
  status: ServiceStatus;
  variant?: string;
};

export const ServiceStatusTag = ({ status, variant }: ServiceStatusTagProps) => {
  const display = serviceStatusToString(status);
  switch (status) {
    case ServiceStatus.RUNNING:
      return (
        <Tag variant={variant} colorScheme={"green"}>
          {display}
        </Tag>
      );
    case ServiceStatus.STOPPED:
      return (
        <Tag variant={variant} colorScheme={"red"}>
          {display}
        </Tag>
      );
    case ServiceStatus.UNKNOWN:
      return (
        <Tag variant={variant} colorScheme={"orange"}>
          {display}
        </Tag>
      );
  }
};
