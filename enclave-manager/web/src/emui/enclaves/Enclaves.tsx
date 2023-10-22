import { Flex } from "@chakra-ui/react";
import { Outlet, RouteObject } from "react-router-dom";
import { Enclave, enclaveLoader, EnclaveLoaderReturnType } from "./Enclave";
import { EnclaveBreadcrumbs } from "./EnclaveBreadcrumbs";
import { EnclaveList } from "./EnclaveList";

export const enclaveRoutes: RouteObject[] = [
  {
    path: "/",
    element: (
      <Flex direction={"column"} gap={"36px"} width={"100%"}>
        <EnclaveBreadcrumbs />
        <Outlet />
      </Flex>
    ),
    handle: { name: () => "Enclaves" },
    children: [
      { path: "/", element: <EnclaveList /> },
      {
        path: "/enclave/:enclaveName",
        element: <Enclave />,
        loader: enclaveLoader,
        handle: { name: (data: EnclaveLoaderReturnType) => data.enclave.name },
      },
    ],
  },
];
