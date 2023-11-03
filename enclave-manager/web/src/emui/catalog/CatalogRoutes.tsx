import { RouteObject } from "react-router-dom";

import { KurtosisPackageIndexerClient } from "../../client/packageIndexer/KurtosisPackageIndexerClient";
import { Catalog } from "./Catalog";
import { catalogLoader } from "./loader";

export const catalogRoutes = (kurtosisIndexerClient: KurtosisPackageIndexerClient): RouteObject[] => [
  {
    path: "/catalog",
    handle: { crumb: () => ({ name: "Catalog", destination: "/catalog" }) },
    loader: catalogLoader(kurtosisIndexerClient),
    id: "catalog",
    element: <Catalog />,
  },
];
