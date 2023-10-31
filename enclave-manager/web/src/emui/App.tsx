import { useMemo } from "react";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { KurtosisClientProvider, useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import {
  KurtosisPackageIndexerProvider,
  useKurtosisPackageIndexerClient,
} from "../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { AppLayout } from "../components/AppLayout";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";
import { catalogRoutes } from "./catalog/CatalogRoutes";
import { enclaveRoutes } from "./enclaves/EnclaveRoutes";
import { Navbar } from "./Navbar";

export const EmuiApp = () => {
  return (
    <KurtosisThemeProvider>
      <KurtosisPackageIndexerProvider>
        <KurtosisClientProvider>
          <KurtosisRouter />
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
      createBrowserRouter([
        {
          element: (
            <AppLayout Nav={<Navbar />}>
              <Outlet />
            </AppLayout>
          ),
          children: [
            { path: "/", children: enclaveRoutes(kurtosisClient) },
            { path: "/catalog", children: catalogRoutes(kurtosisIndexerClient) },
          ],
        },
      ]),
    [kurtosisClient],
  );

  return <RouterProvider router={router} />;
};
