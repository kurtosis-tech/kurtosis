import {
  isDefined,
  KurtosisBreadcrumb,
  KurtosisBreadcrumbsHandle,
  KurtosisBreadcrumbsImpl,
  RemoveFunctions,
} from "kurtosis-ui-components";
import { Fragment, ReactElement, useMemo } from "react";
import { Params, UIMatch } from "react-router-dom";
import { EnclavesState, useEnclavesContext } from "../EnclavesContext";

export type KurtosisEnclavesBreadcrumbsHandle = KurtosisBreadcrumbsHandle<"enclavesHandle"> & {
  crumb?: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => KurtosisBreadcrumb | KurtosisBreadcrumb[];
  extraControls?: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => ReactElement | null;
};

type KurtosisEnclavesBreadcrumbsProps = {
  matches: UIMatch<object, KurtosisEnclavesBreadcrumbsHandle>[];
};

export const KurtosisEnclavesBreadcrumbs = ({ matches }: KurtosisEnclavesBreadcrumbsProps) => {
  const {
    enclaves,
    filesAndArtifactsByEnclave,
    starlarkRunsByEnclave,
    servicesByEnclave,
    starlarkRunningInEnclaves,
    enclaveInitialSubmissionData,
    enclaveInitialPackageId,
  } = useEnclavesContext();

  const matchCrumbs = useMemo(
    () =>
      matches.flatMap((match) => {
        if (isDefined(match.handle?.crumb)) {
          const r = match.handle.crumb(
            {
              enclaves,
              filesAndArtifactsByEnclave,
              starlarkRunsByEnclave,
              servicesByEnclave,
              starlarkRunningInEnclaves,
              enclaveInitialSubmissionData,
              enclaveInitialPackageId,
            },
            match.params,
          );
          return Array.isArray(r) ? r : [r];
        }
        return [];
      }),
    [
      matches,
      enclaves,
      filesAndArtifactsByEnclave,
      starlarkRunsByEnclave,
      servicesByEnclave,
      starlarkRunningInEnclaves,
    ],
  );

  const extraControls = useMemo(
    () =>
      matches
        .map((match) =>
          isDefined(match.handle?.extraControls)
            ? match.handle?.extraControls(
                {
                  enclaves,
                  filesAndArtifactsByEnclave,
                  starlarkRunsByEnclave,
                  servicesByEnclave,
                  starlarkRunningInEnclaves,
                  enclaveInitialSubmissionData,
                  enclaveInitialPackageId,
                },
                match.params,
              )
            : null,
        )
        .filter(isDefined)
        .map((el, i) => <Fragment key={i}>{el}</Fragment>),
    [
      matches,
      enclaves,
      filesAndArtifactsByEnclave,
      starlarkRunsByEnclave,
      servicesByEnclave,
      starlarkRunningInEnclaves,
    ],
  );

  return <KurtosisBreadcrumbsImpl matchCrumbs={matchCrumbs} extraControls={extraControls} />;
};
