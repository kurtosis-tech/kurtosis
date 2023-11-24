import { Icon } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { FiPlus } from "react-icons/fi";
import { Outlet, Params, RouteObject } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { GoToEnclaveOverviewButton } from "../../components/enclaves/GotToEncalaveOverviewButton";
import { KurtosisBreadcrumbsHandle } from "../../components/KurtosisBreadcrumbs";
import { RemoveFunctions } from "../../utils/types";
import { EmuiAppState } from "../EmuiAppContext";
import { Artifact } from "./enclave/artifact/Artifact";
import { Enclave } from "./enclave/Enclave";
import { EnclaveRouteContextProvider } from "./enclave/EnclaveRouteContext";
import { EnclaveLogs } from "./enclave/logs/EnclaveLogs";
import { Service } from "./enclave/service/Service";
import { EnclaveList } from "./EnclaveList";

type KurtosisRouteObject = RouteObject & {
  handle?: KurtosisBreadcrumbsHandle;
  children?: KurtosisRouteObject[];
};

export const enclaveRoutes = (kurtosisClient: KurtosisClient): KurtosisRouteObject[] => [
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
        element: (
          <EnclaveRouteContextProvider>
            <Outlet />
          </EnclaveRouteContextProvider>
        ),
        handle: {
          crumb: ({ enclaves: enclavesResult }: RemoveFunctions<EmuiAppState>, params: Params) => {
            const enclaves = enclavesResult.unwrapOr([]);
            const enclave = enclaves.find((enclave) => enclave.shortenedUuid === params.enclaveUUID);
            return {
              name: enclave?.name || params.enclaveUUID,
              destination: `/enclave/${params.enclaveUUID}`,
              alternatives: [
                ...enclaves
                  .filter((enclave) => enclave.shortenedUuid !== params.enclaveUUID)
                  .sort((a, b) => a.name.localeCompare(b.name))
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
          hasTabs: true,
        },
        children: [
          {
            path: "service/:serviceUUID",
            handle: {
              crumb: ({ servicesByEnclave }: RemoveFunctions<EmuiAppState>, params: Params) => {
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
                    .sort((a, b) => a.name.localeCompare(b.name))
                    .map((service) => ({
                      name: service.name,
                      destination: `/enclave/${params.enclaveUUID}/service/${service.shortenedUuid}`,
                    })),
                };
              },
              hasTabs: true,
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
            element: <Artifact />,
            handle: {
              crumb: ({ filesAndArtifactsByEnclave }: RemoveFunctions<EmuiAppState>, params: Params<string>) => {
                const artifacts = Object.values(
                  filesAndArtifactsByEnclave[params.enclaveUUID || ""]?.unwrapOr({
                    fileNamesAndUuids: [] as FilesArtifactNameAndUuid[],
                  }).fileNamesAndUuids || [],
                );
                const artifact = artifacts.find((artifact) => artifact.fileUuid === params.fileUUID);
                const fileName = artifact?.fileName || "Unknown";

                return [
                  {
                    name: fileName,
                    destination: `/enclave/${params.enclaveUUID}/file/${params.fileUUID}`,
                    alternatives: artifacts
                      .filter((artifact) => artifact.fileUuid !== params.fileUUID)
                      .sort((a, b) => a.fileName.localeCompare(b.fileName))
                      .map((artifact) => ({
                        name: artifact.fileName,
                        destination: `/enclave/${params.enclaveUUID}/file/${artifact.fileUuid}`,
                      })),
                  },
                  { name: "Files", destination: `/enclave/${params.enclaveUUID}/file/${params.fileUUID}` },
                ];
              },
              hasTabs: false,
              extraControls: (state: RemoveFunctions<EmuiAppState>, params: Params<string>) => (
                <GoToEnclaveOverviewButton enclaveUUID={params.enclaveUUID} />
              ),
            },
          },
          {
            path: "logs",
            id: "enclaveLogs",
            element: <EnclaveLogs />,
            handle: {
              hasTabs: false,
              extraControls: ({ starlarkRunningInEnclaves }: RemoveFunctions<EmuiAppState>, params: Params<string>) =>
                starlarkRunningInEnclaves.some((enclave) => enclave.shortenedUuid === params.enclaveUUID) ? null : (
                  <GoToEnclaveOverviewButton enclaveUUID={params.enclaveUUID} />
                ),
              crumb: () => ({
                name: "Logs",
                destination: "none",
              }),
            },
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
