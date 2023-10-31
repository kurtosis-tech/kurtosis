import { Tag } from "@chakra-ui/react";
import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";

export function enclaveStatusToString(status: EnclaveContainersStatus) {
  switch (status) {
    case EnclaveContainersStatus.EnclaveContainersStatus_RUNNING:
      return "Running";
    case EnclaveContainersStatus.EnclaveContainersStatus_STOPPED:
      return "Stopped";
    case EnclaveContainersStatus.EnclaveContainersStatus_EMPTY:
      return "Empty";
  }
}

type EnclaveStatusProps = {
  status: EnclaveContainersStatus;
  variant?: string;
};

export const EnclaveStatus = ({ status, variant }: EnclaveStatusProps) => {
  const display = enclaveStatusToString(status);
  switch (status) {
    case EnclaveContainersStatus.EnclaveContainersStatus_RUNNING:
      return (
        <Tag variant={variant} colorScheme={"green"}>
          {display}
        </Tag>
      );
    case EnclaveContainersStatus.EnclaveContainersStatus_STOPPED:
      return (
        <Tag variant={variant} colorScheme={"red"}>
          {display}
        </Tag>
      );
    case EnclaveContainersStatus.EnclaveContainersStatus_EMPTY:
      return (
        <Tag variant={variant} colorScheme={"gray"}>
          {display}
        </Tag>
      );
  }
};
