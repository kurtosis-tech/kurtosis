import { useMemo } from "react";
import { Params, UIMatch } from "react-router-dom";
import {
  KurtosisBreadcrumb,
  KurtosisBreadcrumbsHandle,
  KurtosisBreadcrumbsImpl,
} from "../../../components/KurtosisBreadcrumbs";
import { isDefined } from "../../../utils";
import { RemoveFunctions } from "../../../utils/types";
import { CatalogState, useCatalogContext } from "../CatalogContext";

export type KurtosisCatalogBreadcrumbsHandle = KurtosisBreadcrumbsHandle<"catalogHandle"> & {
  crumb?: (state: RemoveFunctions<CatalogState>, params: Params<string>) => KurtosisBreadcrumb | KurtosisBreadcrumb[];
};

type KurtosisCatalogBreadcrumbsProps = {
  matches: UIMatch<object, KurtosisCatalogBreadcrumbsHandle>[];
};

export const KurtosisCatalogBreadcrumbs = ({ matches }: KurtosisCatalogBreadcrumbsProps) => {
  const { catalog } = useCatalogContext();

  const matchCrumbs = useMemo(
    () =>
      matches.flatMap((match) => {
        if (isDefined(match.handle?.crumb)) {
          const r = match.handle.crumb({ catalog }, match.params);
          return Array.isArray(r) ? r : [r];
        }
        return [];
      }),
    [matches, catalog],
  );

  return <KurtosisBreadcrumbsImpl matchCrumbs={matchCrumbs} />;
};
