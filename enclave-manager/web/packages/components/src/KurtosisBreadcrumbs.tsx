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
import { ReactElement } from "react";
import { BsCaretDownFill } from "react-icons/bs";
import { Link, UIMatch, useMatches } from "react-router-dom";
import { BREADCRUMBS_HEIGHT } from "./theme/constants";
import { isDefined } from "./utils";

export type KurtosisBreadcrumbsHandle<T extends string> = {
  type: T;
};

type MatchRendererFunction<T extends string> = (props: {
  matches: UIMatch<object, KurtosisBreadcrumbsHandle<T>>[];
}) => ReactElement;

const handlerRegistry: Record<string, MatchRendererFunction<any>> = {};
export const registerBreadcrumbHandler = <T extends string>(
  type: T,
  render: (props: { matches: UIMatch<object, KurtosisBreadcrumbsHandle<T>>[] }) => ReactElement,
) => {
  handlerRegistry[type] = render;
};
const getBreadcumbHandlerRenderer = <T extends string>(type: T): MatchRendererFunction<T> | null => {
  return (handlerRegistry[type] as MatchRendererFunction<T>) || null;
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
  const matches = useMatches() as UIMatch<object, KurtosisBreadcrumbsHandle<any>>[];

  const handlerTypes = new Set(matches.map((match) => match.handle?.type).filter(isDefined));
  if (handlerTypes.size === 0) {
    throw Error(`Currently routes with no breadcrumb handles are not supported`);
  }
  if (handlerTypes.size > 1) {
    throw Error(`Routes with multiple breadcrumb handles are not supported.`);
  }
  const handleType = [...handlerTypes][0];
  const Renderer = getBreadcumbHandlerRenderer(handleType);
  if (isDefined(Renderer)) {
    return <Renderer matches={matches} />;
  }

  throw new Error(`Unable to handle breadcrumbs of type ${handleType}`);
};

type KurtosisBreadcrumbsImplProps = {
  matchCrumbs: KurtosisBreadcrumb[];
  extraControls?: ReactElement[];
};

export const KurtosisBreadcrumbsImpl = ({ matchCrumbs, extraControls }: KurtosisBreadcrumbsImplProps) => {
  return (
    <Flex flex={"none"} h={BREADCRUMBS_HEIGHT} alignItems={"center"} justifyContent={"space-between"}>
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
  );
};

type KurtosisBreadcrumbItemProps = KurtosisBreadcrumb & {
  isLastItem: boolean;
};

const KurtosisBreadcrumbItem = ({ name, destination, alternatives, isLastItem }: KurtosisBreadcrumbItemProps) => {
  const baseLink = isLastItem ? (
    <Text fontSize={"xs"} fontWeight={"semibold"} p={"2px 8px"} borderRadius={"6px"} bg={"gray.650"}>
      {name}
    </Text>
  ) : (
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
