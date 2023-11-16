import { Icon } from "@chakra-ui/react";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FiPlus } from "react-icons/fi";
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
          crumb: async ({ enclaves: enclavesResult }: RemoveFunctions<EmuiAppState>, params: Params) => {
            const enclaves = enclavesResult.unwrapOr([]);
            const enclave = enclaves.find((enclave) => enclave.shortenedUuid === params.enclaveUUID);
            return {
              name: enclave?.name || params.enclaveUUID,
              destination: `/enclave/${params.enclaveUUID}`,
              alternatives: [
                ...enclaves
                  .filter((enclave) => enclave.shortenedUuid !== params.enclaveUUID)
                  .map((enclave) => ({
                    name: enclave.name,
                    destination: `/enclave/${enclave.shortenedUuid}`,
                  })),
                {
                  name: "New Enclave",
                  destination: `${window.location.href}/#create-enclave`,
                  icon: <Icon as={FiPlus} color={"gray.400"} w={"24px"} h={"24px"}></Icon>,
                },
              ],
            };
          },
        },
        children: [
          {
            path: "service/:serviceUUID",
            handle: {
              crumb: async ({ servicesByEnclave }: RemoveFunctions<EmuiAppState>, params: Params) => {
                const services = Object.values(
                  servicesByEnclave[params.enclaveUUID || ""]?.unwrapOr({
                    serviceInfo: {} as Record<string, ServiceInfo>,
                  }).serviceInfo || {},
                );
                const service = services.find((service) => service.shortenedUuid === params.serviceUUID);
                const serviceName = service?.name || "Unknown";

                return {
                  name: serviceName,
                  destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}`,
                  alternatives: services
                    .filter((service) => service.shortenedUuid !== params.serviceUUID)
                    .map((service) => ({
                      name: service.name,
                      destination: `/enclave/${params.enclaveUUID}/service/${service.shortenedUuid}`,
                    })),
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
