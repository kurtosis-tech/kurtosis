import {
  Button,
  Card,
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
import { transportProtocolToString } from "../utils";

type PortsSummaryProps = {
  privatePorts: Record<string, Port>;
  publicPorts: Record<string, Port>;
};

export const PortsSummary = ({ privatePorts, publicPorts }: PortsSummaryProps) => {
  console.log(privatePorts, publicPorts);
  return (
    <Popover trigger={"hover"} preventOverflow isLazy>
      <PopoverTrigger>
        <Button variant="ghost" size="xs">
          {Object.keys(publicPorts).length}
        </Button>
      </PopoverTrigger>
      <PopoverContent maxWidth={"50vw"} w={"unset"}>
        <Flex flexDirection={"row"} gap={"16px"}>
          <Card>
            <PortTable privatePorts={privatePorts} publicPorts={publicPorts} />
          </Card>
        </Flex>
      </PopoverContent>
    </Popover>
  );
};

type PortTableProps = {
  privatePorts: Record<string, Port>;
  publicPorts: Record<string, Port>;
};

const PortTable = ({ privatePorts, publicPorts }: PortTableProps) => {
  if (Object.keys(privatePorts).length === 0) {
    return <i>No ports</i>;
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>Port</Th>
          <Th>Public Port</Th>
          <Th>Application Protocol</Th>
        </Tr>
      </Thead>
      <Tbody>
        {Object.entries(publicPorts)
          .sort(([name1, p1], [name2, p2]) => p1.number - p2.number)
          .map(([name, port], i) => (
            <Tr key={i}>
              <Td>
                {privatePorts[name].number}/{transportProtocolToString(port.transportProtocol)}
              </Td>
              <Td fontSize={"xs"}>{port.number}</Td>
              <Td fontSize={"xs"}>{port.maybeApplicationProtocol || <i>Unknown</i>}</Td>
            </Tr>
          ))}
      </Tbody>
    </Table>
  );
};
