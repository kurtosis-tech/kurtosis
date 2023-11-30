import { Flex, Icon, Image, Text } from "@chakra-ui/react";
import { IoStarSharp } from "react-icons/io5";
import { useKurtosisClient } from "../../client/enclaveManager/KurtosisClientContext";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { RunKurtosisPackageButton } from "./widgets/RunKurtosisPackageButton";
import { SaveKurtosisPackageButton } from "./widgets/SaveKurtosisPackageButton";

type KurtosisPackageCardProps = { kurtosisPackage: KurtosisPackage; onClick?: () => void };

export const KurtosisPackageCard = ({ kurtosisPackage }: KurtosisPackageCardProps) => {
  const client = useKurtosisClient();

  const name = isDefined(kurtosisPackage.repositoryMetadata)
    ? `${kurtosisPackage.repositoryMetadata.name} ${kurtosisPackage.repositoryMetadata.rootPath.split("/").join(" ")}`
    : kurtosisPackage.name;

  return (
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
    >
      <Flex h={"80px"} gap={"16px"} width={"100%"}>
        <Image
          h={"80px"}
          w={"80px"}
          bg={"black"}
          src={`${client.getBaseApplicationUrl()}/logo.png`}
          borderRadius={"6px"}
        />
        <Flex flexDirection={"column"} flex={"1"} justifyContent={"space-between"}>
          <Text noOfLines={2} fontSize={"lg"} textTransform={"capitalize"}>
            {name}
          </Text>
          <Flex justifyContent={"space-between"} fontSize={"xs"}>
            <Text as={"span"} textTransform={"capitalize"}>
              {kurtosisPackage.repositoryMetadata?.owner.replaceAll("-", " ") || "Unknown owner"}
            </Text>
            <Flex gap={"4px"} alignItems={"center"}>
              {kurtosisPackage.stars > 0 && (
                <>
                  <Icon color="gray.500" as={IoStarSharp} />
                  <Text as={"span"}>{kurtosisPackage.stars.toString()}</Text>
                </>
              )}
            </Flex>
          </Flex>
        </Flex>
      </Flex>
      <Flex gap={"16px"} width={"100%"}>
        <SaveKurtosisPackageButton kurtosisPackage={kurtosisPackage} flex={"1"} />
        <RunKurtosisPackageButton kurtosisPackage={kurtosisPackage} flex={"1"} />
      </Flex>
    </Flex>
  );
};
