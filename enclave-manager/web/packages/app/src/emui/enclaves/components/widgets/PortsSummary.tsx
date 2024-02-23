import {
  Flex,
  IconButton,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverContent,
  PopoverTrigger,
  Spinner,
  Tag,
  Text,
} from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { FiChevronDown } from "react-icons/fi";
import { PortsTableRow } from "../tables/PortsTable";
import { PortMaybeLink } from "./PortMaybeLink";

type PortsSummaryProps = {
  ports: "loading" | PortsTableRow[] | null;
};

export const PortsSummary = ({ ports }: PortsSummaryProps) => {
  if (!isDefined(ports)) {
    return <Tag>Unknown</Tag>;
  }

  if (ports === "loading") {
    return <Spinner size={"xs"} />;
  }

  if (ports.length === 0) {
    return (
      <Text fontWeight={"semibold"} fontSize={"xs"} color={"gray.200"}>
        <i>No ports</i>
      </Text>
    );
  }

  const sortedPorts = ports.sort((portA, portB) => {
    if (portA.link.startsWith("http") && portB.link.startsWith("http")) {
      return portA.port.name.localeCompare(portB.port.name);
    }
    if (portA.link.startsWith("http")) {
      return -1;
    }
    if (portB.link.startsWith("http")) {
      return 1;
    }
    return portA.port.name.localeCompare(portB.port.name);
  });
  const priorityPorts = sortedPorts.slice(0, 3);
  const otherPorts = sortedPorts.slice(3);

  return (
    <Flex gap={"4px"} fontWeight={"semibold"} fontSize={"xs"} color={"gray.200"} justifyContent={"center"}>
      {priorityPorts.map((port, i) => (
        <>
          <PortMaybeLink port={port} key={i} />
          {i < priorityPorts.length - 1 && ", "}
        </>
      ))}
      {otherPorts.length > 0 && (
        <Popover>
          <PopoverTrigger>
            <IconButton icon={<FiChevronDown />} variant={"ghost"} size={"xs"} aria-label={"other ports"} />
          </PopoverTrigger>
          <PopoverContent w={"200px"}>
            <PopoverArrow />
            <PopoverBody display={"flex"} flexDirection={"column"}>
              {otherPorts.map((port, i) => (
                <PortMaybeLink port={port} key={i} />
              ))}
            </PopoverBody>
          </PopoverContent>
        </Popover>
      )}
    </Flex>
  );
};
