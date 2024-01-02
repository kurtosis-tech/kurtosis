import { Box, Flex, Icon, Image, Text } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { IoPlay, IoStar } from "react-icons/io5";
import { Link, useHref } from "react-router-dom";
import { readablePackageName } from "./utils";
import { RunKurtosisPackageButton } from "./widgets/RunKurtosisPackageButton";
import { SaveKurtosisPackageButton } from "./widgets/SaveKurtosisPackageButton";

type KurtosisPackageCardProps = {
  kurtosisPackage: KurtosisPackage;
  onRunClick: () => void;
};

export const KurtosisPackageCard = ({ kurtosisPackage, onRunClick }: KurtosisPackageCardProps) => {
  const logoHref = useHref("/logo.png");

  return (
    <Link to={`/catalog/${encodeURIComponent(kurtosisPackage.name)}`}>
      <Flex
        h={"168px"}
        p={"0 24px"}
        bg={"gray.900"}
        borderColor={"whiteAlpha.300"}
        borderWidth={"1px"}
        borderStyle={"solid"}
        borderRadius={"6px"}
        flexDirection={"column"}
        gap={"16px"}
        justifyContent={"center"}
        alignItems={"center"}
        _hover={{ bg: "gray.850", cursor: "pointer" }}
      >
        <Flex h={"80px"} gap={"16px"} width={"100%"}>
          <Image
            h={"80px"}
            w={"80px"}
            bg={kurtosisPackage.iconUrl !== "" ? "white" : "black"}
            src={kurtosisPackage.iconUrl || logoHref}
            fallbackSrc={logoHref}
            borderRadius={"6px"}
          />
          <Flex flexDirection={"column"} flex={"1"} justifyContent={"space-between"}>
            <Text noOfLines={2} fontSize={"lg"}>
              {readablePackageName(kurtosisPackage.name)}
            </Text>
            <Box
              flex={"1"}
              sx={{
                containerType: "size",
                containerName: "details-container",
                "@container details-container (min-height: 30px)": {
                  "> div": { flexDirection: "column", justifyContent: "flex-end", height: "100%" },
                },
              }}
            >
              <Flex justifyContent={"space-between"} fontSize={"xs"} gap={"8px"}>
                <Text as={"span"} textTransform={"capitalize"}>
                  {kurtosisPackage.repositoryMetadata?.owner.replaceAll("-", " ") || "Unknown owner"}
                </Text>
                <Flex gap={"16px"}>
                  <Flex gap={"4px"} alignItems={"center"}>
                    <Icon color="gray.  500" as={IoStar} />
                    <Text as={"span"}>{kurtosisPackage.stars.toString()}</Text>
                  </Flex>
                  <Flex gap={"4px"} alignItems={"center"}>
                    <Icon color="gray.  500" as={IoPlay} />
                    <Text as={"span"}>{kurtosisPackage.runCount.toString()}</Text>
                  </Flex>
                </Flex>
              </Flex>
            </Box>
          </Flex>
        </Flex>
        <Flex gap={"16px"} width={"100%"}>
          <SaveKurtosisPackageButton kurtosisPackage={kurtosisPackage} flex={"1"} />
          <RunKurtosisPackageButton kurtosisPackage={kurtosisPackage} onClick={onRunClick} flex={"1"} />
        </Flex>
      </Flex>
    </Link>
  );
};
