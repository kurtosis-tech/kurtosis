import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Box,
  Card,
  CardBody,
  CardHeader,
  Flex,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  Text,
} from "@chakra-ui/react";
import { useMemo, useRef, useState } from "react";
import { FiSearch } from "react-icons/fi";
import { MdBookmarkAdded } from "react-icons/md";
import { GetPackagesResponse, KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { AppPageLayout } from "../../components/AppLayout";
import { KurtosisPackageCardGrid } from "../../components/catalog/KurtosisPackageCardGrid";
import { OmniboxCommand } from "../../components/KeyboardCommands";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { PageTitle } from "../../components/PageTitle";
import { useKeyboardAction } from "../../components/useKeyboardAction";
import { isDefined } from "../../utils";
import { useCatalogContext } from "./CatalogContext";

export const Catalog = () => {
  const { catalog, savedPackages } = useCatalogContext();

  if (catalog.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={catalog.error} />
      </AppPageLayout>
    );
  }

  return <CatalogImpl savedPackages={savedPackages} catalog={catalog.value} />;
};

type CatalogImplProps = {
  catalog: GetPackagesResponse;
  savedPackages: KurtosisPackage[];
};

const CatalogImpl = ({ catalog, savedPackages }: CatalogImplProps) => {
  const searchRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const isSearching = searchTerm.length > 0;
  const filteredCatalog = useMemo(
    () => catalog.packages.filter((kurtosisPackage) => kurtosisPackage.name.toLowerCase().indexOf(searchTerm) > -1),
    [searchTerm, catalog],
  );

  useKeyboardAction(
    useMemo(
      () => ({
        omniFind: () => {
          if (isDefined(searchRef.current) && searchRef.current !== document.activeElement) {
            searchRef.current.focus();
          }
        },
        escape: () => {
          if (isDefined(searchRef.current) && searchRef.current === document.activeElement) {
            setSearchTerm("");
          }
        },
      }),
      [searchRef],
    ),
  );

  return (
    <AppPageLayout>
      <Flex p={"17px 0"}>
        <PageTitle>Package Catalog</PageTitle>
      </Flex>
      <Flex flexDirection={"column"} gap={"32px"}>
        <Flex flex={"1"} justifyContent={"center"}>
          <InputGroup variant={"solid"} width={"1192px"}>
            <InputLeftElement>
              <Icon as={FiSearch} />
            </InputLeftElement>
            <Input
              ref={searchRef}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder={"Search"}
            />
            <InputRightElement>
              {isSearching ? (
                <IconButton
                  aria-label={"Clear search"}
                  variant="ghost"
                  size={"sm"}
                  icon={<SmallCloseIcon />}
                  onClick={() => setSearchTerm("")}
                />
              ) : (
                <OmniboxCommand />
              )}
            </InputRightElement>
          </InputGroup>
        </Flex>
        {isSearching && (
          <>
            <Heading fontSize={"lg"} fontWeight={"medium"}>
              {filteredCatalog.length} Matches
            </Heading>
            <KurtosisPackageCardGrid packages={filteredCatalog} />
          </>
        )}
        {!isSearching && (
          <>
            {savedPackages.length > 0 && (
              <Box as={"section"} pb="32px" borderColor={"whiteAlpha.300"} borderBottomWidth={"1px"}>
                <Card>
                  <CardHeader
                    display={"flex"}
                    gap={"8px"}
                    alignItems={"center"}
                    fontSize={"lg"}
                    pb={"0"}
                    fontWeight={"medium"}
                  >
                    <Icon as={MdBookmarkAdded} color={"kurtosisGreen.400"} />
                    <Text as={"span"}>Saved</Text>
                  </CardHeader>
                  <CardBody>
                    <KurtosisPackageCardGrid packages={savedPackages} />
                  </CardBody>
                </Card>
              </Box>
            )}
            <Heading fontSize={"lg"} fontWeight={"medium"}>
              All
            </Heading>
            <KurtosisPackageCardGrid packages={catalog.packages} />
          </>
        )}
      </Flex>
    </AppPageLayout>
  );
};
