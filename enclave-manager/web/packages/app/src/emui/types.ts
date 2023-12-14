import { RouteObject } from "react-router-dom";
import { KurtosisCatalogBreadcrumbsHandle } from "./catalog/components/KurtosisCatalogBreadcrumbs";
import { KurtosisEnclavesBreadcrumbsHandle } from "./enclaves/components/KurtosisEnclaveBreadcrumbs";

export type KurtosisEnclavesRouteObject = RouteObject & {
  handle?: KurtosisEnclavesBreadcrumbsHandle;
  children?: KurtosisEnclavesRouteObject[];
};

export type KurtosisCatalogRouteObject = RouteObject & {
  handle?: KurtosisCatalogBreadcrumbsHandle;
  children?: KurtosisCatalogRouteObject[];
};
