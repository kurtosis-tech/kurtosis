import { Box, Flex } from "@chakra-ui/react";
import { AppPageLayout } from "../../components/AppLayout";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { PageTitle } from "../../components/PageTitle";
import { usePackageCatalog } from "./CatalogContext";

export const Catalog = () => {
  const catalog = usePackageCatalog();
  console.log(catalog);

  if (catalog.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={catalog.error} />
      </AppPageLayout>
    );
  }

  return (
    <AppPageLayout>
      <Flex p={"17px 0"}>
        <PageTitle>Package Catalog</PageTitle>
      </Flex>
      <Flex flexDirection={"column"}>
        {catalog.value.packages.map((kurtosisPackage) => (
          <Box key={kurtosisPackage.url}>{kurtosisPackage.name}</Box>
        ))}
      </Flex>
    </AppPageLayout>
  );
};
