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
import { EmuiAppState, useEmuiAppContext } from "../emui/EmuiAppContext";
import { isDefined } from "../utils";
import { RemoveFunctions } from "../utils/types";
import {
  BREADCRUMBS_HEIGHT,
  MAIN_APP_LEFT_PADDING,
  MAIN_APP_MAX_WIDTH_WITHOUT_PADDING,
  MAIN_APP_RIGHT_PADDING,
  MAIN_APP_TOP_PADDING,
} from "./theme/constants";

export type KurtosisBreadcrumbsHandle = {
  crumb?: (state: RemoveFunctions<EmuiAppState>, params: Params<string>) => KurtosisBreadcrumb | KurtosisBreadcrumb[];
  hasTabs?: boolean;
  extraControls?: (state: RemoveFunctions<EmuiAppState>, params: Params<string>) => ReactElement;
};

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
  const { enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave } = useEmuiAppContext();

  const matches = useMatches() as UIMatch<object, KurtosisBreadcrumbsHandle>[];

  const matchCrumbs = useMemo(
    () =>
      matches.flatMap((match) => {
        if (isDefined(match.handle?.crumb)) {
          const r = match.handle.crumb(
            { enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave },
            match.params,
          );
          return Array.isArray(r) ? r : [r];
        }
        return [];
      }),
    [matches, enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave],
  );

  const hasTabs = useMemo(
    () => matches.reduce((acc, match) => (isDefined(match.handle?.hasTabs) ? match.handle.hasTabs : acc), false),
    [matches],
  );

  const extraControls = useMemo(
    () =>
      matches
        .map((match) =>
          isDefined(match.handle?.extraControls)
            ? match.handle?.extraControls(
                { enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave },
                match.params,
              )
            : null,
        )
        .filter(isDefined),
    [matches, enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave],
  );

  return (
    <Flex
      h={BREADCRUMBS_HEIGHT}
      p={`${MAIN_APP_TOP_PADDING} ${MAIN_APP_RIGHT_PADDING} 24px ${MAIN_APP_LEFT_PADDING}`}
      bg={hasTabs ? "gray.850" : undefined}
    >
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
      <Text fontSize={"xs"} fontWeight={"semibold"} color={"gray.400"} p={"0px 8px"}>
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
