import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveFullInfo } from "../../../types";

type ServiceLogsProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

export const ServiceLogs = ({ enclave, service }: ServiceLogsProps) => {
  return <div>Hello world</div>;
};
