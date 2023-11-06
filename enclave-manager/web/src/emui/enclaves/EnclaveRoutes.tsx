import { Params, RouteObject } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { enclavesAction } from "./action";
import { Enclave, enclaveLoader, enclaveTabLoader } from "./enclave";
import { runStarlarkAction } from "./enclave/action";
import { EnclaveLoaderDeferred } from "./enclave/loader";
import { Service } from "./enclave/service/Service";
import { EnclaveList } from "./EnclaveList";
import { enclavesLoader } from "./loader";

export const enclaveRoutes = (kurtosisClient: KurtosisClient): RouteObject[] => [
  {
    path: "/enclaves?",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    loader: enclavesLoader(kurtosisClient),
    action: enclavesAction(kurtosisClient),
    id: "enclaves",
    element: <EnclaveList />,
  },
  {
    path: "/enclave",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    children: [
      {
        path: "/enclave/:enclaveUUID",
        loader: enclaveLoader(kurtosisClient),
        id: "enclave",
        handle: {
          crumb: async (data: EnclaveLoaderDeferred, params: Params) => {
            const resolvedData = await data.data;
            return {
              name: resolvedData.routeName,
              destination: `/enclave/${params.enclaveUUID}`,
            };
          },
        },
        children: [
          {
            path: "service/:serviceUUID/:activeTab?",
            element: <Service />,
          },
          {
            path: "file/:fileUUID",
          },
          {
            path: ":activeTab?",
            loader: enclaveTabLoader,
            action: runStarlarkAction(kurtosisClient),
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
