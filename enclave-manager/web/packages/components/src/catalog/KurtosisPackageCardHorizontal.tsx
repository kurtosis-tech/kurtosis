import { Flex, Icon, Text, Tooltip } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { IoPlay, IoStar } from "react-icons/io5";
import { numberSummary } from "../utils";
import { readablePackageName } from "./utils";
import { PackageLogo } from "./widgets/PackageLogo";
import { SaveKurtosisPackageButton } from "./widgets/SaveKurtosisPackageButton";

type KurtosisPackageCardHorizontalProps = {
  kurtosisPackage: KurtosisPackage;
  onClick: () => void;
};

export const KurtosisPackageCardHorizontal = ({ kurtosisPackage, onClick }: KurtosisPackageCardHorizontalProps) => {
  return (
    <Flex
      h={"76px"}
      w={"100%"}
      p={"16px"}
      bg={"gray.900"}
      borderColor={"whiteAlpha.300"}
      borderWidth={"1px"}
      borderStyle={"solid"}
      borderRadius={"6px"}
      gap={"16px"}
      alignItems={"center"}
      _hover={{ bg: "gray.850", cursor: "pointer" }}
      onClick={onClick}
    >
      <PackageLogo kurtosisPackage={kurtosisPackage} h={"44px"} w={"44px"} />
      <Flex flexDirection={"column"} flex={"1"} justifyContent={"space-between"}>
        <Text noOfLines={1} fontSize={"md"}>
          {readablePackageName(kurtosisPackage.name)}
          <SaveKurtosisPackageButton
            isIconButton
            aria-label={"Toggle button save"}
            kurtosisPackage={kurtosisPackage}
            flex={"1"}
            variant={"ghost"}
          />
        </Text>
        <Flex gap={"12px"} fontSize={"sm"}>
          <Text as={"span"} textTransform={"capitalize"} noOfLines={1} fontSize={"xs"}>
            {kurtosisPackage.repositoryMetadata?.owner.replaceAll("-", " ") || "Unknown owner"}
          </Text>
          <Tooltip label={`This package has ${kurtosisPackage.runCount} stars`}>
            <Flex gap={"4px"} alignItems={"center"}>
              <Icon as={IoStar} />
              <Text as={"span"}>{numberSummary(Number(kurtosisPackage.stars))}</Text>
            </Flex>
          </Tooltip>
          <Tooltip label={`This package has been run ${kurtosisPackage.runCount} times`}>
            <Flex gap={"4px"} alignItems={"center"}>
              <Icon as={IoPlay} />
              <Text as={"span"}>{numberSummary(kurtosisPackage.runCount)}</Text>
            </Flex>
          </Tooltip>
        </Flex>
      </Flex>
    </Flex>
  );
};
