import { readablePackageName, registerBreadcrumbHandler, RemoveFunctions } from "kurtosis-ui-components";
import { Params } from "react-router-dom";
import { KurtosisCatalogRouteObject } from "../types";
import { Catalog } from "./Catalog";
import { CatalogState } from "./CatalogContext";
import { KurtosisCatalogBreadcrumbs } from "./components/KurtosisCatalogBreadcrumbs";
import { Package } from "./package/Package";

registerBreadcrumbHandler("catalogHandle", KurtosisCatalogBreadcrumbs);

export const catalogRoutes = (): KurtosisCatalogRouteObject[] => [
  {
    path: "/catalog/:packageName",
    handle: {
      type: "catalogHandle" as "catalogHandle",
      crumb: ({ catalog }: RemoveFunctions<CatalogState>, params: Params<string>) => {
        const { packageName } = params;
        if (catalog.isErr) {
          return [
            { name: "Catalog", destination: "/catalog" },
            { name: "Unknown", destination: `/catalog/${packageName}` },
          ];
        }

        return [
          { name: "Catalog", destination: "/catalog" },
          { name: readablePackageName(packageName || "Unknown"), destination: `/catalog/${packageName}` },
        ];
      },
    },
    id: "packageDetails",
    element: <Package />,
  },
  {
    path: "/catalog",
    handle: { type: "catalogHandle" as "catalogHandle", crumb: () => ({ name: "Catalog", destination: "/catalog" }) },
    id: "catalog",
    element: <Catalog />,
  },
];
