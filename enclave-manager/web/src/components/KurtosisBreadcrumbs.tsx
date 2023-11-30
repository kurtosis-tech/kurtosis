import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Button,
  ButtonGroup,
  Flex,
  Icon,
  IconButton,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
  Text,
} from "@chakra-ui/react";
import { ReactElement, useMemo } from "react";
import { BsCaretDownFill } from "react-icons/bs";
import { Link, Params, UIMatch, useMatches } from "react-router-dom";
import { EnclavesState, useEnclavesContext } from "../emui/enclaves/EnclavesContext";
import { isDefined } from "../utils";
import { RemoveFunctions } from "../utils/types";
import { BREADCRUMBS_HEIGHT, MAIN_APP_MAX_WIDTH_WITHOUT_PADDING } from "./theme/constants";

type KurtosisBaseBreadcrumbsHandle = {
  type: string;
};

export type KurtosisEnclavesBreadcrumbsHandle = KurtosisBaseBreadcrumbsHandle & {
  type: "enclavesHandle";
  crumb?: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => KurtosisBreadcrumb | KurtosisBreadcrumb[];
  extraControls?: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => ReactElement | null;
};

export type KurtosisCatalogBreadcrumbsHandle = {
  type: "catalogHandle";
  crumb?: () => KurtosisBreadcrumb | KurtosisBreadcrumb[];
};

export type KurtosisBreadcrumbsHandle = KurtosisEnclavesBreadcrumbsHandle | KurtosisCatalogBreadcrumbsHandle;

type KurtosisBreadcrumbMenuItem = {
  name: string;
  destination: string;
  icon?: ReactElement;
};

export type KurtosisBreadcrumb = {
  name: string;
  destination: string;
  alternatives?: KurtosisBreadcrumbMenuItem[];
};

export const KurtosisBreadcrumbs = () => {
  const matches = useMatches() as UIMatch<object, KurtosisBreadcrumbsHandle>[];

  const handlers = new Set(matches.map((match) => match.handle?.type).filter(isDefined));
  if (handlers.size === 0) {
    throw Error(`Currently routes with no breadcrumb handles are not supported`);
  }
  if (handlers.size > 1) {
    throw Error(`Routes with multiple breadcrumb handles are not supported.`);
  }
  const handleType = [...handlers][0];
  const isEnclavesMatches = (
    matches: UIMatch<object, KurtosisBreadcrumbsHandle>[],
    onlyType: KurtosisBreadcrumbsHandle["type"],
  ): matches is UIMatch<object, KurtosisEnclavesBreadcrumbsHandle>[] => onlyType === "enclavesHandle";
  const isCatalogMatches = (
    matches: UIMatch<object, KurtosisBreadcrumbsHandle>[],
    onlyType: KurtosisBreadcrumbsHandle["type"],
  ): matches is UIMatch<object, KurtosisCatalogBreadcrumbsHandle>[] => onlyType === "catalogHandle";
  if (isEnclavesMatches(matches, handleType)) {
    return <KurtosisEnclavesBreadcrumbs matches={matches} />;
  }
  if (isCatalogMatches(matches, handleType)) {
    return <KurtosisCatalogBreadcrumbs matches={matches} />;
  }

  throw new Error(`Unable to handle breadcrumbs of type ${handleType}`);
};

type KurtosisEnclavesBreadcrumbsProps = {
  matches: UIMatch<object, KurtosisEnclavesBreadcrumbsHandle>[];
};

const KurtosisEnclavesBreadcrumbs = ({ matches }: KurtosisEnclavesBreadcrumbsProps) => {
  const { enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave, starlarkRunningInEnclaves } =
    useEnclavesContext();

  const matchCrumbs = useMemo(
    () =>
      matches.flatMap((match) => {
        if (isDefined(match.handle?.crumb)) {
          const r = match.handle.crumb(
            {
              enclaves,
              filesAndArtifactsByEnclave,
              starlarkRunsByEnclave,
              servicesByEnclave,
              starlarkRunningInEnclaves,
            },
            match.params,
          );
          return Array.isArray(r) ? r : [r];
        }
        return [];
      }),
    [
      matches,
      enclaves,
      filesAndArtifactsByEnclave,
      starlarkRunsByEnclave,
      servicesByEnclave,
      starlarkRunningInEnclaves,
    ],
  );

  const extraControls = useMemo(
    () =>
      matches
        .map((match) =>
          isDefined(match.handle?.extraControls)
            ? match.handle?.extraControls(
                {
                  enclaves,
                  filesAndArtifactsByEnclave,
                  starlarkRunsByEnclave,
                  servicesByEnclave,
                  starlarkRunningInEnclaves,
                },
                match.params,
              )
            : null,
        )
        .filter(isDefined),
    [
      matches,
      enclaves,
      filesAndArtifactsByEnclave,
      starlarkRunsByEnclave,
      servicesByEnclave,
      starlarkRunningInEnclaves,
    ],
  );

  return <KurtosisBreadcrumbsImpl matchCrumbs={matchCrumbs} extraControls={extraControls} />;
};

type KurtosisCatalogBreadcrumbsProps = {
  matches: UIMatch<object, KurtosisCatalogBreadcrumbsHandle>[];
};

const KurtosisCatalogBreadcrumbs = ({ matches }: KurtosisCatalogBreadcrumbsProps) => {
  const matchCrumbs = useMemo(
    () =>
      matches.flatMap((match) => {
        if (isDefined(match.handle?.crumb)) {
          const r = match.handle.crumb();
          return Array.isArray(r) ? r : [r];
        }
        return [];
      }),
    [matches],
  );

  return <KurtosisBreadcrumbsImpl matchCrumbs={matchCrumbs} />;
};

type KurtosisBreadcrumbsImplProps = {
  matchCrumbs: KurtosisBreadcrumb[];
  extraControls?: ReactElement[];
};

const KurtosisBreadcrumbsImpl = ({ matchCrumbs, extraControls }: KurtosisBreadcrumbsImplProps) => {
  return (
    <Flex h={BREADCRUMBS_HEIGHT}>
      <Flex w={MAIN_APP_MAX_WIDTH_WITHOUT_PADDING} alignItems={"center"} justifyContent={"space-between"}>
        <Flex>
          <Breadcrumb
            variant={"topNavigation"}
            separator={
              <Text as={"span"} fontSize={"lg"}>
                /
              </Text>
            }
          >
            <BreadcrumbItem>
              <Text fontSize={"xs"} fontWeight={"semibold"} p={"0px 8px"}>
                Kurtosis
              </Text>
            </BreadcrumbItem>
            {matchCrumbs.map((crumb, i, arr) => (
              <BreadcrumbItem key={i} isCurrentPage={i === arr.length - 1}>
                <KurtosisBreadcrumbItem {...crumb} key={i} isLastItem={i === arr.length - 1} />
              </BreadcrumbItem>
            ))}
          </Breadcrumb>
          &nbsp;
        </Flex>
        <Flex>{extraControls}</Flex>
      </Flex>
    </Flex>
  );
};

type KurtosisBreadcrumbItemProps = KurtosisBreadcrumb & {
  isLastItem: boolean;
};

const KurtosisBreadcrumbItem = ({ name, destination, alternatives, isLastItem }: KurtosisBreadcrumbItemProps) => {
  if (isLastItem) {
    return (
      <Text fontSize={"xs"} fontWeight={"semibold"} p={"2px 8px"} borderRadius={"6px"} bg={"gray.650"}>
        {name}
      </Text>
    );
  }

  const baseLink = (
    <BreadcrumbLink as={Link} to={destination}>
      <Button variant={"breadcrumb"} size={"xs"}>
        {name}
      </Button>
    </BreadcrumbLink>
  );

  if (isDefined(alternatives) && alternatives.length > 0) {
    // Render with menu
    return (
      <ButtonGroup isAttached>
        {baseLink}
        <Menu>
          <MenuButton
            as={IconButton}
            variant={"breadcrumb"}
            aria-label={"Other options"}
            icon={<Icon as={BsCaretDownFill} />}
            size={"xs"}
          />
          <MenuList>
            {alternatives.map(({ name, destination, icon }) => (
              <MenuItem key={destination} as={Link} to={destination} icon={icon}>
                {name}
              </MenuItem>
            ))}
          </MenuList>
        </Menu>
      </ButtonGroup>
    );
  }
  return baseLink;
};
