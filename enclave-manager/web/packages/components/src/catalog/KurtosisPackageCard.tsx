import { Box, Flex, Icon, Text, Tooltip } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { IoPlay, IoStar } from "react-icons/io5";
import { Link } from "react-router-dom";
import { numberSummary } from "../utils";
import { readablePackageName } from "./utils";
import { PackageLogo } from "./widgets/PackageLogo";
import { RunKurtosisPackageButton } from "./widgets/RunKurtosisPackageButton";
import { SaveKurtosisPackageButton } from "./widgets/SaveKurtosisPackageButton";

type KurtosisPackageCardProps = {
  kurtosisPackage: KurtosisPackage;
  onRunClick: () => void;
};

export const KurtosisPackageCard = ({ kurtosisPackage, onRunClick }: KurtosisPackageCardProps) => {
  return (
    <Link to={`/catalog/${encodeURIComponent(kurtosisPackage.name)}`}>
      <Flex
        position={"relative"}
        h={"260px"}
        w={"204px"}
        p={"24px"}
        bg={"gray.900"}
        borderColor={"whiteAlpha.300"}
        borderWidth={"1px"}
        borderStyle={"solid"}
        borderRadius={"6px"}
        flexDirection={"column"}
        gap={"16px"}
        justifyContent={"space-between"}
        alignItems={"center"}
        _hover={{ bg: "gray.850", cursor: "pointer" }}
      >
        <Box position={"absolute"} top={"8px"} right={"8px"}>
          <SaveKurtosisPackageButton
            isIconButton
            aria-label={"Toggle button save"}
            kurtosisPackage={kurtosisPackage}
            flex={"1"}
            variant={"ghost"}
          />
        </Box>
        <PackageLogo logoUrl={kurtosisPackage.iconUrl} h={"80px"} w={"80px"} />
        <Flex h={"80px"} gap={"8px"} width={"100%"}>
          <Flex flexDirection={"column"} flex={"1"} justifyContent={"space-between"}>
            <Text noOfLines={1} fontSize={"md"} fontWeight={"bold"}>
              {readablePackageName(kurtosisPackage.name)}
            </Text>
            <Text as={"span"} textTransform={"capitalize"} noOfLines={1} fontSize={"xs"}>
              {kurtosisPackage.repositoryMetadata?.owner.replaceAll("-", " ") || "Unknown owner"}
            </Text>
            <Flex gap={"16px"} color="gray.200" fontSize={"xs"}>
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
        <Flex gap={"16px"} width={"100%"}>
          <RunKurtosisPackageButton kurtosisPackage={kurtosisPackage} onClick={onRunClick} flex={"1"} />
        </Flex>
      </Flex>
    </Link>
  );
};
