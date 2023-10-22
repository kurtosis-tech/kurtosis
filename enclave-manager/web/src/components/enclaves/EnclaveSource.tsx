import { Flex, Icon } from "@chakra-ui/react";
import { IoLogoGithub } from "react-icons/io";

type EnclaveSourceProps = {
  source: string;
};

export const EnclaveSource = ({ source }: EnclaveSourceProps) => {
  if (source.startsWith("github.com/")) {
    return (
      <Flex gap={"6px"} alignItems={"center"}>
        <Icon as={IoLogoGithub} h={"24px"} w={"24px"} color={"kurtosisGray.500"}></Icon>
        <span>{source.replace("github.com/", "")}</span>
      </Flex>
    );
  }

  return <Flex>{source}</Flex>;
};
