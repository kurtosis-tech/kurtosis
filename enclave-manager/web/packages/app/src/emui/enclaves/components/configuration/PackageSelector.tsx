import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Button,
  Divider,
  Flex,
  Icon,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  Spinner,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
} from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import {
  FindCommand,
  isDefined,
  KurtosisAlert,
  KurtosisPackageCardHorizontal,
  parsePackageUrl,
  useKeyboardAction,
  useSavedPackages,
} from "kurtosis-ui-components";
import { debounce } from "lodash";
import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { FiSearch } from "react-icons/fi";
import { useCatalogContext } from "../../../catalog/CatalogContext";

type ExactMatchState =
  | { type: "loading"; url: string }
  | { type: "loaded"; package: KurtosisPackage }
  | { type: "error"; error: string };

type PackageSelectorProps = {
  onPackageSelected: (kurtosisPackage: KurtosisPackage) => void;
};
export const PackageSelector = ({ onPackageSelected }: PackageSelectorProps) => {
  const searchRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState("");

  const [exactMatch, setExactMatch] = useState<ExactMatchState>();
  const { catalog, getSinglePackage } = useCatalogContext();

  const checkSinglePackageDebounced = useMemo(
    () =>
      debounce(
        async (packageName: string) => {
          const singlePackageResult = await getSinglePackage(packageName);
          if (singlePackageResult.isErr) {
            setExactMatch({ type: "error", error: singlePackageResult.error });
            return;
          }
          if (isDefined(singlePackageResult.value.package)) {
            setExactMatch({ type: "loaded", package: singlePackageResult.value.package });
          }
        },
        1000,
        { trailing: true, leading: false },
      ),
    [getSinglePackage],
  );

  const startCheckSinglePackage = useCallback(
    async (packageName: string) => {
      let isKurtosisPackageName = false;
      try {
        parsePackageUrl(packageName);
        isKurtosisPackageName = true;
      } catch (error: any) {
        // This packageName isn't a kurtosis package url
      }
      if (isKurtosisPackageName) {
        setExactMatch({ type: "loading", url: packageName });
        checkSinglePackageDebounced(packageName);
      } else {
        setExactMatch(undefined);
      }
    },
    [checkSinglePackageDebounced],
  );

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

  useKeyboardAction(useMemo(() => ({ find: () => searchRef.current?.focus() }), [searchRef]));

  useEffect(() => {
    startCheckSinglePackage(searchTerm);
  }, [startCheckSinglePackage, searchTerm]);

  if (searchResults.isErr) {
    return <KurtosisAlert message={"Unable to load kurtosis packages"} details={searchResults.error} />;
  }

  return (
    <Tabs>
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
      <Flex p={2}></Flex>
      <TabList>
        <Tab>Kurtosis Packages</Tab>
        <Tab>Docker-compose Packages</Tab>
      </TabList>
      <Divider />
      <Flex p={4}></Flex>
      <TabPanels>
        <TabPanel>
          {isDefined(exactMatch) && (
            <Flex flexDirection={"column"} gap={"10px"}>
              <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
                Exact Match
              </Text>
              {exactMatch.type === "loading" && (
                <Flex flexDirection={"column"} alignItems={"center"}>
                  <Spinner />
                  <Text>Looking for a Kurtosis Package at {exactMatch.url}</Text>
                </Flex>
              )}
              {exactMatch.type === "loaded" && (
                <KurtosisPackageCardHorizontal
                  kurtosisPackage={exactMatch.package}
                  onClick={() => onPackageSelected(exactMatch.package)}
                />
              )}
              {exactMatch.type === "error" && (
                <KurtosisAlert message={"Error looking up package"} details={exactMatch.error} />
              )}
            </Flex>
          )}
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
              <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
                All Packages
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
        </TabPanel>
        <TabPanel>
          <p>Docker compose!</p>
        </TabPanel>
      </TabPanels>
    </Tabs>
  );
};
