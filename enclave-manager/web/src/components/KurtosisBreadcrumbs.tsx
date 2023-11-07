import { ChevronRightIcon } from "@chakra-ui/icons";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, Flex } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { Link, Params, UIMatch, useMatches } from "react-router-dom";
import { isDefined } from "../utils";

export type KurtosisBreadcrumb = {
  name: string;
  destination: string;
};

export const KurtosisBreadcrumbs = () => {
  const matches = useMatches() as UIMatch<
    object,
    {
      crumb?: (
        data: Record<string, object>,
        params: Params<string>,
      ) => KurtosisBreadcrumb | Promise<KurtosisBreadcrumb>;
    }
  >[];

  const [matchCrumbs, setMatchCrumbs] = useState<KurtosisBreadcrumb[]>([]);

  useEffect(() => {
    (async () => {
      const allLoaderData = matches
        .filter((match) => isDefined(match.data))
        .reduce((acc, match) => ({ ...acc, [match.id]: match.data }), {});

      setMatchCrumbs(
        await Promise.all(
          matches
            .map((match) =>
              isDefined(match.handle?.crumb) ? Promise.resolve(match.handle.crumb(allLoaderData, match.params)) : null,
            )
            .filter(isDefined),
        ),
      );
    })();
  }, [matches]);

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
