import { useMemo } from "react";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { KurtosisClientProvider, useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import {
  KurtosisPackageIndexerProvider,
  useKurtosisPackageIndexerClient,
} from "../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { AppLayout } from "../components/AppLayout";
import { CreateEnclave } from "../components/enclaves/CreateEnclave";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";
import { catalogRoutes } from "./catalog/CatalogRoutes";
import { EmuiAppContextProvider } from "./EmuiAppContext";
import { enclaveRoutes } from "./enclaves/EnclaveRoutes";
import { Navbar } from "./Navbar";

export const EmuiApp = () => {
  return (
    <KurtosisThemeProvider>
      <KurtosisPackageIndexerProvider>
        <KurtosisClientProvider>
          <EmuiAppContextProvider>
            <KurtosisRouter />
          </EmuiAppContextProvider>
        </KurtosisClientProvider>
      </KurtosisPackageIndexerProvider>
    </KurtosisThemeProvider>
  );
};

const KurtosisRouter = () => {
  const kurtosisClient = useKurtosisClient();
  const kurtosisIndexerClient = useKurtosisPackageIndexerClient();

  const router = useMemo(
    () =>
      createBrowserRouter(
        [
          {
            element: (
              <AppLayout Nav={<Navbar />}>
                <Outlet />
                <CreateEnclave />
              </AppLayout>
            ),
            children: [
              { path: "/", children: enclaveRoutes(kurtosisClient) },
              { path: "/catalog", children: catalogRoutes(kurtosisIndexerClient) },
            ],
          },
        ],
        {
          basename: kurtosisClient.getBaseApplicationUrl().pathname,
        },
      ),
    [kurtosisClient, kurtosisIndexerClient],
  );

  return <RouterProvider router={router} />;
};
