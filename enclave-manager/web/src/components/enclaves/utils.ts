import { Port_TransportProtocol } from "enclave-manager-sdk/build/api_container_service_pb";

export function transportProtocolToString(protocol: Port_TransportProtocol) {
  switch (protocol) {
    case Port_TransportProtocol.TCP:
      return "TCP";
    case Port_TransportProtocol.SCTP:
      return "SCTP";
    case Port_TransportProtocol.UDP:
      return "UDP";
  }
}
