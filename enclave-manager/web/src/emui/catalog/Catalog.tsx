import { Box, Flex } from "@chakra-ui/react";
import { Suspense } from "react";
import { Await, useLoaderData } from "react-router-dom";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { CatalogLoaderResolved } from "./loader";

export const Catalog = () => {
  const { catalog } = useLoaderData() as CatalogLoaderResolved;

  return (
    <Suspense>
      <Await resolve={catalog} children={(catalog) => <CatalogImpl catalog={catalog} />} />
    </Suspense>
  );
};

type CatalogImplProps = {
  catalog: CatalogLoaderResolved["catalog"];
};

const CatalogImpl = ({ catalog }: CatalogImplProps) => {
  if (catalog.isErr) {
    return <KurtosisAlert message={catalog.error} />;
  }

  return (
    <Flex flexDirection={"column"}>
      {catalog.value.map((kurtosisPackage) => (
        <Box key={kurtosisPackage.url}>{kurtosisPackage.name}</Box>
      ))}
    </Flex>
  );
};
