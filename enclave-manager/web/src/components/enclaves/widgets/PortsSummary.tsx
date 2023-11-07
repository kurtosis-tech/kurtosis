import {
  Button,
  Flex,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@chakra-ui/react";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { TitledCard } from "../../TitledCard";
import { transportProtocolToString } from "../utils";

type PortsSummaryProps = {
  privatePorts: Port[];
  publicPorts: Port[];
};

export const PortsSummary = ({ privatePorts, publicPorts }: PortsSummaryProps) => {
  return (
    <Popover trigger={"hover"} preventOverflow isLazy>
      <PopoverTrigger>
        <Button variant="ghost" size="xs">
          {privatePorts.length + publicPorts.length}
        </Button>
      </PopoverTrigger>
      <PopoverContent maxWidth={"50vw"} w={"unset"}>
        <Flex flexDirection={"row"} gap={"16px"}>
          <TitledCard title={"Public Ports"}>
            <PortTable ports={publicPorts} />
          </TitledCard>
          <TitledCard title={"Private Ports"}>
            <PortTable ports={privatePorts} />
          </TitledCard>
        </Flex>
      </PopoverContent>
    </Popover>
  );
};

type PortTableProps = {
  ports: Port[];
};

const PortTable = ({ ports }: PortTableProps) => {
  if (ports.length === 0) {
    return <i>No ports</i>;
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>Port</Th>
          <Th>Protocol</Th>
          <Th>Application Protocol</Th>
          <Th>Timeout</Th>
        </Tr>
      </Thead>
      <Tbody>
        {ports
          .sort((p1, p2) => p1.number - p2.number)
          .map((port, i) => (
            <Tr key={i}>
              <Td>{port.number}</Td>
              <Td fontSize={"xs"}>{transportProtocolToString(port.transportProtocol)}</Td>
              <Td fontSize={"xs"}>{port.maybeApplicationProtocol || <i>Unknown</i>}</Td>
              <Td fontSize={"xs"}>{port.maybeWaitTimeout || ""}</Td>
            </Tr>
          ))}
      </Tbody>
    </Table>
  );
};
