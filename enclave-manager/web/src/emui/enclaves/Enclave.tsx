import { Box } from "@chakra-ui/react";
import { UIMatch } from "@remix-run/router";
import { LoaderFunction, useLoaderData, useMatches } from "react-router-dom";

export type EnclaveLoaderData = {
  name: string;
};

export type EnclaveLoaderReturnType = { enclave: EnclaveLoaderData };

export type EnclaveRouteHandles = {
  name: (data?: EnclaveLoaderReturnType) => string;
};

export const useEnclaveRouteMatches = () => {
  return useMatches() as UIMatch<EnclaveLoaderReturnType, EnclaveRouteHandles>[];
};

export const enclaveLoader: LoaderFunction = async ({ params }): Promise<EnclaveLoaderReturnType> => {
  return { enclave: { name: params.enclaveName || "Unknown" } };
};

export const Enclave = () => {
  const { enclave } = useLoaderData() as EnclaveLoaderReturnType;
  return <Box>Enclave {enclave.name}</Box>;
};
