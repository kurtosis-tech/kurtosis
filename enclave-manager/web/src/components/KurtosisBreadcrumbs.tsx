import { ChevronRightIcon } from "@chakra-ui/icons";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, Flex } from "@chakra-ui/react";
import { Link, Params, UIMatch, useMatches } from "react-router-dom";
import { isDefined } from "../utils";

export type KurtosisBreadcrumb = {
  name: string;
  destination: string;
};

export const KurtosisBreadcrumbs = () => {
  const matches = useMatches() as UIMatch<
    object,
    { crumb?: (data: object, params: Params<string>) => KurtosisBreadcrumb }
  >[];

  const matchCrumbs = matches
    .map((match) => (isDefined(match.handle?.crumb) ? match.handle.crumb(match.data, match.params) : null))
    .filter(isDefined);

  return (
    <Flex h="40px" p={"4px 0"} alignItems={"center"}>
      <Breadcrumb variant={"topNavigation"} separator={<ChevronRightIcon h={"20px"} w={"24px"} />}>
        {matchCrumbs.map(({ name, destination }, i, arr) => (
          <BreadcrumbItem key={i} isCurrentPage={i === arr.length - 1}>
            <BreadcrumbLink as={i === arr.length - 1 ? undefined : Link} to={destination}>
              {name}
            </BreadcrumbLink>
          </BreadcrumbItem>
        ))}
      </Breadcrumb>
      &nbsp;
    </Flex>
  );
};
