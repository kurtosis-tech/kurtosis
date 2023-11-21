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
import { ReactElement, useEffect, useState } from "react";
import { BsCaretDownFill } from "react-icons/bs";
import { Link, Params, UIMatch, useMatches } from "react-router-dom";
import { EmuiAppState, useEmuiAppContext } from "../emui/EmuiAppContext";
import { isDefined } from "../utils";
import { RemoveFunctions } from "../utils/types";
import { MAIN_APP_LEFT_PADDING, MAIN_APP_RIGHT_PADDING, MAIN_APP_TOP_PADDING } from "./theme/constants";

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

  const matches = useMatches() as UIMatch<
    object,
    {
      crumb?: (
        state: RemoveFunctions<EmuiAppState>,
        params: Params<string>,
      ) => KurtosisBreadcrumb | Promise<KurtosisBreadcrumb>;
    }
  >[];

  const [matchCrumbs, setMatchCrumbs] = useState<KurtosisBreadcrumb[]>([]);

  useEffect(() => {
    (async () => {
      setMatchCrumbs(
        await Promise.all(
          matches
            .map((match) =>
              isDefined(match.handle?.crumb)
                ? Promise.resolve(
                    match.handle.crumb(
                      { enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave },
                      match.params,
                    ),
                  )
                : null,
            )
            .filter(isDefined),
        ),
      );
    })();
  }, [matches, enclaves, filesAndArtifactsByEnclave, starlarkRunsByEnclave, servicesByEnclave]);

  return (
    <Flex
      h={"76px"}
      p={`${MAIN_APP_TOP_PADDING} ${MAIN_APP_RIGHT_PADDING} 24px ${MAIN_APP_LEFT_PADDING}`}
      alignItems={"center"}
      bg={"gray.850"}
    >
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
