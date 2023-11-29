import { RouteObject } from "react-router-dom";
import { KurtosisCatalogBreadcrumbsHandle, KurtosisEnclavesBreadcrumbsHandle } from "../components/KurtosisBreadcrumbs";

export type KurtosisEnclavesRouteObject = RouteObject & {
  handle?: KurtosisEnclavesBreadcrumbsHandle;
  children?: KurtosisEnclavesRouteObject[];
};

export type KurtosisCatalogRouteObject = RouteObject & {
  handle?: KurtosisCatalogBreadcrumbsHandle;
  children?: KurtosisEnclavesRouteObject[];
};
