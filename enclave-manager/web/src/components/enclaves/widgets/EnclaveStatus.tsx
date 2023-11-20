import { Tag, Tooltip } from "@chakra-ui/react";
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

export function enclaveStatusToColorScheme(status: EnclaveContainersStatus) {
  switch (status) {
    case EnclaveContainersStatus.EnclaveContainersStatus_RUNNING:
      return "green";
    case EnclaveContainersStatus.EnclaveContainersStatus_STOPPED:
      return "red";
    case EnclaveContainersStatus.EnclaveContainersStatus_EMPTY:
      return "gray";
  }
}

type EnclaveStatusProps = {
  status: EnclaveContainersStatus;
  variant?: string;
};

export const EnclaveStatus = ({ status, variant }: EnclaveStatusProps) => {
  const display = enclaveStatusToString(status);
  const colorScheme = enclaveStatusToColorScheme(status);

  return (
    <Tooltip closeDelay={1000} label={"This is the status of the container running the enclave"}>
      <Tag variant={variant} colorScheme={colorScheme}>
        {display}
      </Tag>
    </Tooltip>
  );
};
