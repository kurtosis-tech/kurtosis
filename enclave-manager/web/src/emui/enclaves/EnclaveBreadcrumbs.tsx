import { Breadcrumbs, EnclaveBreadCrumb } from "../../components/Breadcrumbs";
import { isDefined } from "../../utils";
import { useEnclaveRouteMatches } from "./types";

export const EnclaveBreadcrumbs = () => {
  const matches = useEnclaveRouteMatches();

  const matchCrumbs = matches
    .filter((match) => isDefined(match.handle?.name))
    .map((match) => ({ name: match.handle.name(match.data), destination: match.pathname }));

  const crumbs: EnclaveBreadCrumb[] = [
    { name: "Kurtosis Enclave Manager", destination: "/" },
    ...(matchCrumbs.length > 1 ? matchCrumbs : []),
  ];

  return <Breadcrumbs crumbs={crumbs} />;
};
