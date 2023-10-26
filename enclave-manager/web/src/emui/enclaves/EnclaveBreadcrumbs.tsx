import { Flex } from "@chakra-ui/react";
import { KurtosisBreadcrumb, KurtosisBreadcrumbs } from "../../components/KurtosisBreadcrumbs";
import { isDefined } from "../../utils";
import { useEnclaveRouteMatches } from "./Enclave";

export const EnclaveBreadcrumbs = () => {
  const matches = useEnclaveRouteMatches();

  const matchCrumbs = matches
    .filter((match) => isDefined(match.handle?.name))
    .map((match) => ({ name: match.handle.name(match.data), destination: match.pathname }));

  const crumbs: KurtosisBreadcrumb[] = [...(matchCrumbs.length > 1 ? matchCrumbs : [])];

  return (
    <Flex h="40px" p={"4px 0"} flexDirection={"column"} justifyContent={"center"} alignItems={"flex-start"}>
      <KurtosisBreadcrumbs crumbs={crumbs} />
      &nbsp;
    </Flex>
  );
};
