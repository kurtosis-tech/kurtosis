import { Icon } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined, registerBreadcrumbHandler, RemoveFunctions } from "kurtosis-ui-components";
import { FiPlus } from "react-icons/fi";
import { Outlet, Params } from "react-router-dom";
import { KurtosisEnclavesRouteObject } from "../types";
import { GoToEnclaveOverviewButton } from "./components/GotToEncalaveOverviewButton";
import { KurtosisEnclavesBreadcrumbs } from "./components/KurtosisEnclaveBreadcrumbs";
import { KurtosisUpgrader } from "./components/KurtosisUpgrader";
import { Artifact } from "./enclave/artifact/Artifact";
import { Enclave } from "./enclave/Enclave";
import { EnclaveRouteContextProvider } from "./enclave/EnclaveRouteContext";
import { EnclaveLogs } from "./enclave/logs/EnclaveLogs";
import { Service } from "./enclave/service/Service";
import { EnclaveList } from "./EnclaveList";
import { EnclavesState } from "./EnclavesContext";

registerBreadcrumbHandler("enclavesHandle", KurtosisEnclavesBreadcrumbs);

export const enclaveRoutes = (): KurtosisEnclavesRouteObject[] => [
  {
    path: "/enclaves?",
    handle: {
      type: "enclavesHandle" as "enclavesHandle",
      crumb: () => ({ name: "Enclaves", destination: "/" }),
      extraControls: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => <KurtosisUpgrader />,
    },
    id: "enclaves",
    element: <EnclaveList />,
  },
  {
    path: "/enclave",
    handle: { type: "enclavesHandle" as "enclavesHandle", crumb: () => ({ name: "Enclaves", destination: "/" }) },
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
          type: "enclavesHandle" as "enclavesHandle",
          crumb: ({ enclaves: enclavesResult }: RemoveFunctions<EnclavesState>, params: Params) => {
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
              type: "enclavesHandle" as "enclavesHandle",
              crumb: ({ servicesByEnclave }: RemoveFunctions<EnclavesState>, params: Params) => {
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
                  type: "enclavesHandle" as "enclavesHandle",
                  crumb: (data: RemoveFunctions<EnclavesState>, params: Params<string>) => {
                    const activeTab = params.activeTab;

                    if (!isDefined(activeTab) || activeTab.toLowerCase() === "overview") {
                      return [];
                    }

                    return {
                      name: "Logs",
                      destination: `/enclave/${params.enclaveUUID}/service/${params.serviceUUID}/logs`,
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
              type: "enclavesHandle" as "enclavesHandle",
              crumb: ({ filesAndArtifactsByEnclave }: RemoveFunctions<EnclavesState>, params: Params<string>) => {
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
              extraControls: (state: RemoveFunctions<EnclavesState>, params: Params<string>) => (
                <GoToEnclaveOverviewButton enclaveUUID={params.enclaveUUID} />
              ),
            },
          },
          {
            path: "logs",
            id: "enclaveLogs",
            element: <EnclaveLogs />,
            handle: {
              type: "enclavesHandle" as "enclavesHandle",
              hasTabs: false,
              extraControls: ({ starlarkRunningInEnclaves }: RemoveFunctions<EnclavesState>, params: Params<string>) =>
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
              type: "enclavesHandle" as "enclavesHandle",
              crumb: (data: RemoveFunctions<EnclavesState>, params: Params<string>) => {
                const activeTab = params.activeTab;

                if (!isDefined(activeTab) || activeTab.toLowerCase() === "overview") {
                  return [];
                }

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
