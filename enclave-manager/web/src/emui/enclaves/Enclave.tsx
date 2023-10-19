import { LoaderFunction, useLoaderData } from "react-router-dom";
import { Box } from "@chakra-ui/react";
import { EnclaveLoaderReturnType } from "./types";

export const enclaveLoader: LoaderFunction = async ({ params }): Promise<EnclaveLoaderReturnType> => {
  return { enclave: { name: params.enclaveName || "Unknown" } };
};

export const Enclave = () => {
  const { enclave } = useLoaderData() as EnclaveLoaderReturnType;
  return <Box>Enclave {enclave.name}</Box>;
};
