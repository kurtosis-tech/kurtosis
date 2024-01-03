import { SmallCloseIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  ButtonGroup,
  Card,
  CardBody,
  Checkbox,
  Flex,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  Menu,
  MenuButton,
  MenuDivider,
  MenuItem,
  MenuList,
} from "@chakra-ui/react";
import { GetPackagesResponse, KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import {
  AppPageLayout,
  FindCommand,
  isDefined,
  KurtosisAlert,
  KurtosisPackageCardGrid,
  KurtosisPackageCardRow,
  maybeArrayToArray,
  PageTitle,
  useKeyboardAction,
  useSavedPackages,
} from "kurtosis-ui-components";
import { useEffect, useMemo, useRef, useState } from "react";
import { BiSortAlt2 } from "react-icons/bi";
import { FiSearch } from "react-icons/fi";
import { HiStar } from "react-icons/hi";
import { IoFilterSharp, IoPlay } from "react-icons/io5";
import { MdBookmarkAdded } from "react-icons/md";
import { useSearchParams } from "react-router-dom";
import { ConfigureEnclaveModal } from "../enclaves/components/modals/ConfigureEnclaveModal";
import { EnclavesContextProvider } from "../enclaves/EnclavesContext";
import { useCatalogContext } from "./CatalogContext";

type SearchState = {
  term: string;
  filter: ("saved" | "featured")[]; // TODO: Implement 'featured'
  sortBy?: "stars" | "runs";
};

export const Catalog = () => {
  const { catalog } = useCatalogContext();
  const { savedPackages } = useSavedPackages();

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
  const [searchParams, setUrlSearchParams] = useSearchParams();
  const initialTerm = searchParams.get("t") || "";
  const initialFilter = maybeArrayToArray(searchParams.get("f")).filter(isDefined);
  const initialSortBy = searchParams.get("s") || undefined;

  const [searchTerm, setSearchTerm] = useState<SearchState>({
    term: initialTerm,
    filter: initialFilter as ("saved" | "featured")[],
    sortBy: initialSortBy as "stars" | "runs",
  });
  const [configuringPackage, setConfiguringPackage] = useState<KurtosisPackage>();
  const isSearching = searchTerm.term.length > 0 || searchTerm.filter.length > 0 || isDefined(searchTerm.sortBy);
  const filteredCatalog = useMemo(
    () =>
      catalog.packages
        .filter((kurtosisPackage) => kurtosisPackage.name.toLowerCase().indexOf(searchTerm.term) > -1)
        .filter((kurtosisPackage) => {
          if (searchTerm.filter.length === 0) {
            return true;
          }
          if (
            searchTerm.filter.indexOf("saved") >= 0 &&
            savedPackages.some((savedKurtosisPackage) => savedKurtosisPackage.name === kurtosisPackage.name)
          ) {
            return true;
          }
          // TODO: Implement 'featured' filtering
          return false;
        })
        .sort((a, b) => {
          if (searchTerm.sortBy === "stars") {
            return a.stars > b.stars ? -1 : a.stars === b.stars ? 0 : 1;
          }
          if (searchTerm.sortBy === "runs") {
            return b.runCount - a.runCount;
          }
          return 0;
        }),
    [searchTerm, catalog, savedPackages],
  );

  const mostStarredPackages = useMemo(
    () => [...catalog.packages].sort((a, b) => (a.stars > b.stars ? -1 : a.stars === b.stars ? 0 : 1)).slice(0, 10),
    [catalog],
  );
  const mostRanPackages = useMemo(
    () =>
      [...catalog.packages]
        .sort((a, b) => (a.runCount > b.runCount ? -1 : a.runCount === b.runCount ? 0 : 1))
        .slice(0, 10),
    [catalog],
  );

  const handlePackageRun = (kurtosisPackage: KurtosisPackage) => {
    setConfiguringPackage(kurtosisPackage);
  };

  const handleFilterToggle = (filter: "saved" | "featured") => () => {
    setSearchTerm((searchTerm) => ({
      ...searchTerm,
      filter:
        searchTerm.filter.indexOf(filter) >= 0
          ? searchTerm.filter.filter((searchFilter) => searchFilter !== filter)
          : [...searchTerm.filter, filter],
    }));
  };

  useKeyboardAction(
    useMemo(
      () => ({
        find: () => {
          if (isDefined(searchRef.current) && searchRef.current !== document.activeElement) {
            searchRef.current.focus();
          }
        },
        escape: () => {
          if (isDefined(searchRef.current) && searchRef.current === document.activeElement) {
            setSearchTerm((searchTerm) => ({ ...searchTerm, term: "" }));
          }
        },
      }),
      [searchRef],
    ),
  );

  useEffect(() => {
    const params = new URLSearchParams();
    params.set("t", searchTerm.term);
    searchTerm.filter.forEach((f) => params.set("f", f));
    if (isDefined(searchTerm.sortBy)) {
      params.set("s", searchTerm.sortBy);
    }
    const currentParts = window.location.href.split("?");
    if (currentParts.length > 1) {
      setUrlSearchParams(params, { replace: true });
    } else {
      setUrlSearchParams(params);
    }
  }, [searchTerm, setUrlSearchParams]);

  return (
    <AppPageLayout>
      <Flex p={"17px 0"}>
        <PageTitle>Package Catalog</PageTitle>
      </Flex>
      <Flex flexDirection={"column"} gap={"32px"}>
        <Flex flex={"1"} justifyContent={"center"}>
          <InputGroup variant={"solid"} width={"1192px"} color={"gray.150"}>
            <InputLeftElement>
              <Icon as={FiSearch} />
            </InputLeftElement>
            <Input
              ref={searchRef}
              value={searchTerm.term}
              bgColor={"gray.850"}
              onChange={(e) => setSearchTerm((searchTerm) => ({ ...searchTerm, term: e.target.value }))}
              placeholder={"Search"}
            />
            <InputRightElement w={"unset"}>
              {isSearching ? (
                <IconButton
                  aria-label={"Clear search"}
                  variant="ghost"
                  size={"sm"}
                  icon={<SmallCloseIcon />}
                  onClick={() => setSearchTerm((searchTerm) => ({ filter: [], term: "" }))}
                />
              ) : (
                <FindCommand whiteSpace={"nowrap"} pr={"10px"} />
              )}
            </InputRightElement>
          </InputGroup>
        </Flex>
        {isSearching && (
          <Flex flexDirection={"column"} gap={"32px"} maxW={"1248px"}>
            <Flex justifyContent={"space-between"} alignItems={"center"}>
              <Heading fontSize={"lg"} fontWeight={"medium"}>
                {filteredCatalog.length} Matches
              </Heading>
              <ButtonGroup variant={"ghost"} size={"xs"}>
                <Menu closeOnSelect={false}>
                  <MenuButton as={Button} leftIcon={<IoFilterSharp />}>
                    Filter
                  </MenuButton>
                  <MenuList>
                    <MenuItem>
                      <Flex as={"span"} gap={"8px"}>
                        <Checkbox
                          onChange={handleFilterToggle("saved")}
                          isChecked={searchTerm.filter.indexOf("saved") >= 0}
                        >
                          Saved
                        </Checkbox>
                      </Flex>
                    </MenuItem>
                    <MenuDivider />
                    <MenuItem onClick={() => setSearchTerm((searchTerm) => ({ ...searchTerm, filter: [] }))}>
                      Clear
                    </MenuItem>
                  </MenuList>
                </Menu>
                <Menu>
                  <MenuButton as={Button} rightIcon={<BiSortAlt2 />}>
                    Sort
                  </MenuButton>
                  <MenuList>
                    <MenuItem
                      icon={<HiStar />}
                      onClick={() => setSearchTerm((searchTerm) => ({ ...searchTerm, sortBy: "stars" }))}
                    >
                      Stars
                    </MenuItem>
                    <MenuItem
                      icon={<IoPlay />}
                      onClick={() => setSearchTerm((searchTerm) => ({ ...searchTerm, sortBy: "runs" }))}
                    >
                      Run Count
                    </MenuItem>
                  </MenuList>
                </Menu>
              </ButtonGroup>
            </Flex>
            <KurtosisPackageCardGrid packages={filteredCatalog} onPackageRunClicked={handlePackageRun} />
          </Flex>
        )}
        {!isSearching && (
          <>
            {savedPackages.length > 0 && (
              <Box as={"section"} pb="32px" borderColor={"whiteAlpha.300"} borderBottomWidth={"1px"}>
                <Card>
                  <CardBody>
                    <KurtosisPackageCardRow
                      title={"Saved"}
                      icon={<Icon as={MdBookmarkAdded} color={"kurtosisGreen.400"} />}
                      packages={savedPackages}
                      onPackageRunClicked={handlePackageRun}
                      onSeeAllClicked={() => setSearchTerm((searchTerm) => ({ ...searchTerm, filter: ["saved"] }))}
                    />
                  </CardBody>
                </Card>
              </Box>
            )}
            <Box as={"section"} pb="32px" borderColor={"whiteAlpha.300"} borderBottomWidth={"1px"}>
              <KurtosisPackageCardRow
                title={"Most starred"}
                packages={mostStarredPackages}
                onPackageRunClicked={handlePackageRun}
                onSeeAllClicked={() => setSearchTerm((searchTerm) => ({ ...searchTerm, sortBy: "stars" }))}
              />
            </Box>
            <Box as={"section"} pb="32px" borderColor={"whiteAlpha.300"} borderBottomWidth={"1px"}>
              <KurtosisPackageCardRow
                title={"Most Ran"}
                packages={mostRanPackages}
                onPackageRunClicked={handlePackageRun}
                onSeeAllClicked={() => setSearchTerm((searchTerm) => ({ ...searchTerm, sortBy: "runs" }))}
              />
            </Box>
            <Heading fontSize={"lg"} fontWeight={"medium"}>
              All
            </Heading>
            <KurtosisPackageCardGrid packages={catalog.packages} onPackageRunClicked={handlePackageRun} />
          </>
        )}
        {configuringPackage && (
          <EnclavesContextProvider skipInitialLoad>
            <ConfigureEnclaveModal
              isOpen={true}
              onClose={() => setConfiguringPackage(undefined)}
              kurtosisPackage={configuringPackage}
            />
          </EnclavesContextProvider>
        )}
      </Flex>
    </AppPageLayout>
  );
};
