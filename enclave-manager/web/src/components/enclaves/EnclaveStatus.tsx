import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";
import { Tag } from "@chakra-ui/react";

type EnclaveStatusProps = {
  status: EnclaveContainersStatus;
};

export const EnclaveStatus = ({ status }: EnclaveStatusProps) => {
  switch (status) {
    case EnclaveContainersStatus.EnclaveContainersStatus_RUNNING:
      return (
        <Tag variant={"kurtosisSubtle"} colorScheme={"green"}>
          RUNNING
        </Tag>
      );
    case EnclaveContainersStatus.EnclaveContainersStatus_STOPPED:
      return (
        <Tag variant={"kurtosisSubtle"} colorScheme={"red"}>
          Stopped
        </Tag>
      );
    case EnclaveContainersStatus.EnclaveContainersStatus_EMPTY:
      return (
        <Tag variant={"kurtosisSubtle"} colorScheme={"gray"}>
          Empty
        </Tag>
      );
  }
};
