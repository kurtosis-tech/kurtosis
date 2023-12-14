import { Port_TransportProtocol } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "kurtosis-ui-components";

export function transportProtocolToString(protocol: Port_TransportProtocol) {
  switch (protocol) {
    case Port_TransportProtocol.TCP:
      return "TCP";
    case Port_TransportProtocol.SCTP:
      return "SCTP";
    case Port_TransportProtocol.UDP:
      return "UDP";
    default:
      return "";
  }
}

export const allowedEnclaveNamePattern = /^[-A-Za-z0-9]{1,60}$/;

export function isEnclaveNameAllowed(name: any): boolean {
  if (typeof name !== "string") {
    return false;
  }
  return isDefined(name.match(allowedEnclaveNamePattern));
}
