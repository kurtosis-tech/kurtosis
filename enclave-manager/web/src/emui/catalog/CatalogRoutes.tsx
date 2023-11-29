import { KurtosisCatalogRouteObject } from "../types";
import { Catalog } from "./Catalog";

export const catalogRoutes = (): KurtosisCatalogRouteObject[] => [
  {
    path: "/catalog",
    handle: { type: "catalogHandle" as "catalogHandle", crumb: () => ({ name: "Catalog", destination: "/catalog" }) },
    id: "catalog",
    element: <Catalog />,
  },
];
