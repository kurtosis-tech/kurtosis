import { Params, RouteObject } from "react-router-dom";
import { Enclave, enclaveLoader, enclaveTabLoader } from "./Enclave";
import { EnclaveList, enclavesLoader } from "./EnclaveList";

export const enclaveRoutes: RouteObject[] = [
  {
    path: "",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    loader: enclavesLoader,
    id: "enclaves",
    children: [
      { path: "", element: <EnclaveList /> },
      {
        path: "enclave/:uuid",
        loader: enclaveLoader,
        handle: {
          crumb: (data: Awaited<ReturnType<typeof enclaveLoader>>, params: Params<string>) => ({
            name: data.routeName,
            destination: `/enclave/${params.uuid}`,
          }),
        },
        children: [
          {
            path: ":activeTab?",
            loader: enclaveTabLoader,
            element: <Enclave />,
            handle: {
              crumb: (data: Awaited<ReturnType<typeof enclaveTabLoader>>, params: Params<string>) => ({
                name: data.routeName,
                destination: `/enclave/${params.uuid}/${params.activeTab || "overview"}`,
              }),
            },
          },
        ],
      },
    ],
  },
];
