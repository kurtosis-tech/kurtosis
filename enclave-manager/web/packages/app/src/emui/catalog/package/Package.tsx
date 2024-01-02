import { Box, Flex, Icon, Image, Input, InputGroup, InputRightElement, Text } from "@chakra-ui/react";
import { DateTime } from "luxon";
import { useState } from "react";
import { IoPlay, IoStar } from "react-icons/io5";
import { useParams } from "react-router-dom";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";

import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import {
  AppPageLayout,
  CopyButton,
  FormatDateTime,
  isDefined,
  KurtosisAlert,
  KurtosisMarkdown,
  PackageSourceButton,
  readablePackageName,
  RunKurtosisPackageButton,
  SaveKurtosisPackageButton,
  TitledCard,
} from "kurtosis-ui-components";
import { ConfigureEnclaveModal } from "../../enclaves/components/modals/ConfigureEnclaveModal";
import { EnclavesContextProvider } from "../../enclaves/EnclavesContext";
import { useKurtosisPackage } from "../CatalogContext";

export const Package = () => {
  const { packageName } = useParams();
  const kurtosisPackage = useKurtosisPackage(packageName || "unknown");

  if (kurtosisPackage.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={kurtosisPackage.error} />
      </AppPageLayout>
    );
  }

  return <PackageImpl kurtosisPackage={kurtosisPackage.value} />;
};

type PackageImplProps = {
  kurtosisPackage: KurtosisPackage;
};

const PackageImpl = ({ kurtosisPackage }: PackageImplProps) => {
  const [showConfigurePackage, setShowConfigurePackage] = useState(false);
  const runCommand = `kurtosis run ${kurtosisPackage.name}`;

  return (
    <AppPageLayout>
      <Flex flexDirection={"column"} width={"100%"} h={"100%"} gap={"32px"}>
        <PackageHeader kurtosisPackage={kurtosisPackage} />
        <Flex gap={"32px"} h={"100%"} w={"100%"}>
          <Flex gap={"32px"} flexDirection={"column"} maxW={"740px"} w={"100%"}>
            <TitledCard title={"DESCRIPTION"}>
              <Box p={"0 15px 15px 15px"}>
                <KurtosisMarkdown>{kurtosisPackage.description}</KurtosisMarkdown>
              </Box>
            </TitledCard>
          </Flex>
          <Flex flexDirection={"column"} gap={"16px"} flex={"1"}>
            <RunKurtosisPackageButton
              kurtosisPackage={kurtosisPackage}
              size={"lg"}
              onClick={() => setShowConfigurePackage(true)}
            />
            <InputGroup size={"lg"}>
              <Input value={runCommand} textOverflow={"ellipsis"} />
              <InputRightElement>
                <CopyButton
                  contentName={"command"}
                  isIconButton
                  aria-label={"Click to copy this command"}
                  valueToCopy={runCommand}
                />
              </InputRightElement>
            </InputGroup>
            <PackageSourceButton
              source={kurtosisPackage.name}
              variant={"outline"}
              color={"gray.100"}
              size={"lg"}
              width={"100%"}
            >
              View on Github
            </PackageSourceButton>
            <Flex borderBottomWidth={"1px"} borderBottomColor={"whiteAlpha.300"} gap={"32px"} p={"16px"}>
              <Flex gap={"16px"} flexDirection={"column"}>
                <Flex gap={"8px"} color="gray.400" fontWeight={"bold"} alignItems={"center"}>
                  <Icon as={IoStar} w={"12px"} h={"12px"} />
                  <Text as={"span"} textTransform={"uppercase"}>
                    Stars
                  </Text>
                </Flex>
                <Text as={"span"} fontWeight={"medium"} fontSize={"xl"}>
                  {kurtosisPackage.stars.toString()}
                </Text>
              </Flex>
              <Flex gap={"16px"} flexDirection={"column"}>
                <Flex gap={"8px"} color="gray.400" fontWeight={"bold"} alignItems={"center"}>
                  <Icon as={IoPlay} w={"12px"} h={"12px"} />
                  <Text as={"span"} textTransform={"uppercase"}>
                    Run Count
                  </Text>
                </Flex>
                <Text as={"span"} fontWeight={"medium"} fontSize={"xl"}>
                  {kurtosisPackage.runCount.toString()}
                </Text>
              </Flex>
            </Flex>
            <Flex gap={"16px"} flexDirection={"column"} p={"16px"}>
              <Text as={"span"} color="gray.400" fontWeight={"bold"} textTransform={"uppercase"}>
                Last updated
              </Text>
              <span>
                <FormatDateTime
                  fontWeight={"medium"}
                  fontSize={"xl"}
                  format={"relative"}
                  flex={"0 1 auto"}
                  dateTime={
                    isDefined(kurtosisPackage.repositoryMetadata) &&
                    isDefined(kurtosisPackage.repositoryMetadata.lastCommitTime)
                      ? DateTime.fromJSDate(kurtosisPackage.repositoryMetadata.lastCommitTime.toDate())
                      : null
                  }
                />
              </span>
            </Flex>
          </Flex>
        </Flex>
        {showConfigurePackage && (
          <EnclavesContextProvider skipInitialLoad>
            <ConfigureEnclaveModal
              isOpen={true}
              onClose={() => setShowConfigurePackage(false)}
              kurtosisPackage={kurtosisPackage}
            />
          </EnclavesContextProvider>
        )}
      </Flex>
    </AppPageLayout>
  );
};

const PackageHeader = ({ kurtosisPackage }: PackageImplProps) => {
  const client = useKurtosisClient();

  return (
    <Flex gap={"22px"} w={"100%"}>
      <Image
        h={"120px"}
        w={"120px"}
        bg={kurtosisPackage.iconUrl !== "" ? "white" : "black"}
        src={kurtosisPackage.iconUrl || `${client.getBaseApplicationUrl()}/logo.png`}
        fallbackSrc={`${client.getBaseApplicationUrl()}/logo.png`}
        borderRadius={"9px"}
      />
      <Flex flexDirection={"column"} justifyContent={"space-between"} flex={"1"}>
        <Flex flexDirection={"column"} gap={"8px"}>
          <Text noOfLines={1} fontSize={"xl"}>
            {readablePackageName(kurtosisPackage.name)}
          </Text>
          <Text as={"span"} textTransform={"capitalize"}>
            {kurtosisPackage.repositoryMetadata?.owner.replaceAll("-", " ") || "Unknown owner"}
          </Text>
        </Flex>
        <Box>
          <SaveKurtosisPackageButton kurtosisPackage={kurtosisPackage} />
        </Box>
      </Flex>
    </Flex>
  );
};
