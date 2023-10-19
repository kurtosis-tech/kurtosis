import { useMatches } from "react-router-dom";
import { UIMatch } from "@remix-run/router";

export type EnclaveData = {
  name: string;
};

export type EnclaveLoaderReturnType = { enclave: EnclaveData };

export type EnclaveRouteHandles = {
  name: (data?: EnclaveLoaderReturnType) => string;
};

export const useEnclaveRouteMatches = () => {
  return useMatches() as UIMatch<EnclaveLoaderReturnType, EnclaveRouteHandles>[];
};
