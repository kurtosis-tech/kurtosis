import { Breadcrumbs, EnclaveBreadCrumb } from "../../components/Breadcrumbs";
import { isDefined } from "../../utils";
import { useEnclaveRouteMatches } from "./types";
import { Flex } from "@chakra-ui/react";

export const EnclaveBreadcrumbs = () => {
  const matches = useEnclaveRouteMatches();

  const matchCrumbs = matches
    .filter((match) => isDefined(match.handle?.name))
    .map((match) => ({ name: match.handle.name(match.data), destination: match.pathname }));

  const crumbs: EnclaveBreadCrumb[] = [...(matchCrumbs.length > 1 ? matchCrumbs : [])];

  return (
    <Flex h="40px" p={"4px 0"} flexDirection={"column"} justifyContent={"center"} alignItems={"flex-start"}>
      <Breadcrumbs crumbs={crumbs} />
      &nbsp;
    </Flex>
  );
};
