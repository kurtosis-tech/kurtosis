import { Params, RouteObject } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { enclavesAction } from "./action";
import { Enclave, enclaveLoader, enclaveTabLoader } from "./enclave";
import { EnclaveList } from "./EnclaveList";
import { enclavesLoader } from "./loader";

export const enclaveRoutes = (kurtosisClient: KurtosisClient): RouteObject[] => [
  {
    path: "/enclaves?",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    loader: enclavesLoader(kurtosisClient),
    action: enclavesAction(kurtosisClient),
    id: "enclaves",
    children: [
      { path: "", element: <EnclaveList /> },
      {
        path: "enclave/:enclaveUUID",
        loader: enclaveLoader(kurtosisClient),
        handle: {
          crumb: (data: Awaited<ReturnType<ReturnType<typeof enclaveLoader>>>, params: Params<string>) => ({
            name: data.routeName,
            destination: `/enclave/${params.enclaveUUID}`,
          }),
        },
        children: [
          {
            path: "service/:serviceUUID",
          },
          {
            path: "file/:fileUUID",
          },
          {
            path: ":activeTab?",
            loader: enclaveTabLoader,
            element: <Enclave />,
            handle: {
              crumb: (data: Awaited<ReturnType<typeof enclaveTabLoader>>, params: Params<string>) => ({
                name: data.routeName,
                destination: `/enclave/${params.enclaveUUID}/${params.activeTab || "overview"}`,
              }),
            },
          },
        ],
      },
    ],
  },
];
