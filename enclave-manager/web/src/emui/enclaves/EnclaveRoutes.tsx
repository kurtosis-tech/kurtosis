import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { Params, RouteObject } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { RemoveFunctions } from "../../utils/types";
import { EmuiAppState } from "../EmuiAppContext";
import { Enclave } from "./enclave/Enclave";
import { Service } from "./enclave/service/Service";
import { EnclaveList } from "./EnclaveList";

export const enclaveRoutes = (kurtosisClient: KurtosisClient): RouteObject[] => [
  {
    path: "/enclaves?",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    id: "enclaves",
    element: <EnclaveList />,
  },
  {
    path: "/enclave",
    handle: { crumb: () => ({ name: "Enclaves", destination: "/" }) },
    children: [
      {
        path: "/enclave/:enclaveUUID",
        id: "enclave",
        handle: {
          crumb: async ({ enclaves }: RemoveFunctions<EmuiAppState>, params: Params) => {
            const enclave = enclaves.unwrapOr([]).find((enclave) => enclave.shortenedUuid === params.enclaveUUID);
            return {
              name: enclave?.name || params.enclaveUUID,
              destination: `/enclave/${params.enclaveUUID}`,
            };
          },
        },
        children: [
          {
            path: "service/:serviceUUID",
            handle: {
              crumb: async ({ servicesByEnclave }: RemoveFunctions<EmuiAppState>, params: Params) => {
                const service = Object.values(
                  servicesByEnclave[params.enclaveUUID || ""]?.unwrapOr({
                    serviceInfo: {} as Record<string, ServiceInfo>,
                  }).serviceInfo || {},
                ).find((service) => service.shortenedUuid === params.serviceUUID);
                const serviceName = service?.name || "Unknown";

                return {
                  name: serviceName,
                  destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}`,
                };
              },
            },
            children: [
              {
                path: ":activeTab?",
                id: "serviceActiveTab",
                element: <Service />,
                handle: {
                  crumb: (data: RemoveFunctions<EmuiAppState>, params: Params<string>) => {
                    const activeTab = params.activeTab;

                    let routeName = activeTab?.toLowerCase() === "logs" ? "Logs" : "Overview";

                    return {
                      name: routeName,
                      destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}/${
                        params.activeTab || "overview"
                      }`,
                    };
                  },
                },
              },
            ],
          },
          {
            path: "file/:fileUUID",
          },
          {
            path: ":activeTab?",
            id: "enclaveActiveTab",
            element: <Enclave />,
            handle: {
              crumb: (data: RemoveFunctions<EmuiAppState>, params: Params<string>) => {
                const activeTab = params.activeTab;

                let routeName =
                  activeTab?.toLowerCase() === "logs"
                    ? "Logs"
                    : activeTab?.toLowerCase() === "source"
                    ? "Source"
                    : "Overview";

                return {
                  name: routeName,
                  destination: `/enclave/${params.enclaveUUID}/${params.activeTab || "overview"}`,
                };
              },
            },
          },
        ],
      },
    ],
  },
];
