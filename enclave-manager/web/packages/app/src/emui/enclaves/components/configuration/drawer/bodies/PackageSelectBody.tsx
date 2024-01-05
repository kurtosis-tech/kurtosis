import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Button,
  DrawerBody,
  DrawerFooter,
  DrawerHeader,
  Flex,
  Icon,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  Text,
} from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { FindCommand, KurtosisAlert, KurtosisPackageCardHorizontal, useSavedPackages } from "kurtosis-ui-components";
import { ChangeEvent, useMemo, useRef, useState } from "react";
import { FiSearch } from "react-icons/fi";
import { useCatalogContext } from "../../../../../catalog/CatalogContext";
import { DrawerExpandCollapseButton } from "../DrawerExpandCollapseButton";
import { DrawerSizes } from "../types";

type PackageSelectBodyProps = {
  onPackageSelected: (kurtosisPackage: KurtosisPackage) => void;
  onClose: () => void;
  drawerSize: DrawerSizes;
  onDrawerSizeClick: () => void;
};
export const PackageSelectBody = ({
  onPackageSelected,
  onClose,
  drawerSize,
  onDrawerSizeClick,
}: PackageSelectBodyProps) => {
  const searchRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState("");

  const { catalog } = useCatalogContext();

  const searchResults = useMemo(
    () =>
      catalog.map((catalog) =>
        catalog.packages.filter(
          (kurtosisPackage) => kurtosisPackage.name.toLowerCase().indexOf(searchTerm.toLowerCase()) >= 0,
        ),
      ),
    [catalog, searchTerm],
  );

  const { savedPackages } = useSavedPackages();

  const handleSearchTermChange = async (e: ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  if (searchResults.isErr) {
    return (
      <DrawerBody>
        <KurtosisAlert message={"Unable to load kurtosis packages"} details={searchResults.error} />
      </DrawerBody>
    );
  }

  return (
    <>
      <DrawerHeader display={"flex"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
        <DrawerExpandCollapseButton drawerSize={drawerSize} onClick={onDrawerSizeClick} />
        <Text as={"span"}>Enclave Configuration</Text>
        {/*Here to balance the space-between*/}
        <Text />
      </DrawerHeader>
      <DrawerBody>
        <InputGroup variant={"solid"} width={"100%"} color={"gray.150"}>
          <InputLeftElement>
            <Icon as={FiSearch} />
          </InputLeftElement>
          <Input
            ref={searchRef}
            value={searchTerm}
            bgColor={"gray.850"}
            onChange={handleSearchTermChange}
            placeholder={"Search"}
            autoFocus
          />
          <InputRightElement w={"unset"} mr={"8px"}>
            {searchTerm.length > 0 ? (
              <Button variant="ghost" size={"xs"} rightIcon={<SmallCloseIcon />} onClick={() => setSearchTerm("")}>
                Clear
              </Button>
            ) : (
              <FindCommand whiteSpace={"nowrap"} pr={"10px"} />
            )}
          </InputRightElement>
        </InputGroup>
        {(searchTerm.length > 0 || savedPackages.length === 0) && (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
              {searchTerm.length === 0 ? "All Packages" : "Search Results"}
            </Text>
            {searchResults.value.map((kurtosisPackage) => (
              <KurtosisPackageCardHorizontal
                key={kurtosisPackage.name}
                kurtosisPackage={kurtosisPackage}
                onClick={() => onPackageSelected(kurtosisPackage)}
              />
            ))}
          </Flex>
        )}
        {searchTerm.length === 0 && savedPackages.length > 0 && (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
              Saved
            </Text>
            {savedPackages.map((kurtosisPackage) => (
              <KurtosisPackageCardHorizontal
                key={kurtosisPackage.name}
                kurtosisPackage={kurtosisPackage}
                onClick={() => onPackageSelected(kurtosisPackage)}
              />
            ))}
          </Flex>
        )}
      </DrawerBody>
      <DrawerFooter>
        <Flex justifyContent={"space-between"} gap={"12px"} width={"100%"}>
          <Button color={"gray.100"} onClick={onClose}>
            Cancel
          </Button>
        </Flex>
      </DrawerFooter>
    </>
  );
};
