import { ChevronRightIcon } from "@chakra-ui/icons";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, Button, Flex } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { Link, Params, UIMatch, useMatches } from "react-router-dom";
import { EmuiAppState, useEmuiAppContext } from "../emui/EmuiAppContext";
import { isDefined } from "../utils";
import { RemoveFunctions } from "../utils/types";

export type KurtosisBreadcrumb = {
  name: string;
  destination: string;
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
    <Flex h="40px" p={"4px 0"} alignItems={"center"}>
      <Breadcrumb variant={"topNavigation"} separator={<ChevronRightIcon h={"20px"} w={"24px"} />}>
        {matchCrumbs.map(({ name, destination }, i, arr) => (
          <BreadcrumbItem key={i} isCurrentPage={i === arr.length - 1}>
            <BreadcrumbLink as={i === arr.length - 1 ? undefined : Link} to={destination}>
              {i === arr.length - 1 ? (
                name
              ) : (
                <Button variant={"breadcrumb"} size={"sm"}>
                  {name}
                </Button>
              )}
            </BreadcrumbLink>
          </BreadcrumbItem>
        ))}
      </Breadcrumb>
      &nbsp;
    </Flex>
  );
};
