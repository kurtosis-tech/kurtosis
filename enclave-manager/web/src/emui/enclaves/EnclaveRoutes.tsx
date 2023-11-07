import { Params, RouteObject } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { enclavesAction } from "./action";
import { Enclave, enclaveLoader, enclaveTabLoader } from "./enclave";
import { runStarlarkAction } from "./enclave/action";
import { EnclaveLoaderDeferred } from "./enclave/loader";
import { Service } from "./enclave/service/Service";
import { EnclaveList } from "./EnclaveList";
import { enclavesLoader } from "./loader";
import { serviceTabLoader } from "./enclave/service/tabLoader";

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
          crumb: async (data: Record<string, object>, params: Params) => {
            const resolvedData = await (data["enclave"] as EnclaveLoaderDeferred).data;
            return {
              name: resolvedData.routeName,
              destination: `/enclave/${params.enclaveUUID}`,
            };
          },
        },
        children: [
          {
            path: "service/:serviceUUID",
            handle: {
              crumb: async (data: Record<string, object>, params: Params) => {
                const resolvedData = await (data["enclave"] as EnclaveLoaderDeferred).data;
                let serviceName = "Unknown";
                if (
                  resolvedData.enclave &&
                  resolvedData.enclave.isOk &&
                  resolvedData.enclave.value.services.isOk &&
                  params.serviceUUID
                ) {
                  const service = Object.values(resolvedData.enclave.value.services.value.serviceInfo).find(
                    (service) => service.shortenedUuid === params.serviceUUID,
                  );
                  if (service) {
                    serviceName = service.name;
                  }
                }

                return {
                  name: serviceName,
                  destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}`,
                };
              },
            },
            children: [
              {
                path: ":activeTab?",
                loader: serviceTabLoader,
                id: "serviceActiveTab",
                element: <Service />,
                handle: {
                  crumb: (data: Record<string, object>, params: Params<string>) => ({
                    name: (data["serviceActiveTab"] as Awaited<ReturnType<typeof serviceTabLoader>>).routeName,
                    destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}/${
                      params.activeTab || "overview"
                    }`,
                  }),
                },
              },
            ],
          },
          {
            path: "file/:fileUUID",
          },
          {
            path: ":activeTab?",
            loader: enclaveTabLoader,
            action: runStarlarkAction(kurtosisClient),
            id: "enclaveActiveTab",
            element: <Enclave />,
            handle: {
              crumb: (data: Record<string, object>, params: Params<string>) => ({
                name: (data["enclaveActiveTab"] as Awaited<ReturnType<typeof enclaveTabLoader>>).routeName,
                destination: `/enclave/${params.enclaveUUID}/${params.activeTab || "overview"}`,
              }),
            },
          },
        ],
      },
    ],
  },
];
