import { AppLayout, KurtosisThemeProvider } from "kurtosis-ui-components";
import { useMemo } from "react";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { KurtosisClientProvider, useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import { KurtosisPackageIndexerProvider } from "../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { CatalogContextProvider } from "./catalog/CatalogContext";
import { catalogRoutes } from "./catalog/CatalogRoutes";
import { BuildEnclave } from "./enclaves/components/BuildEnclave";
import { CreateEnclave } from "./enclaves/components/CreateEnclave";
import { enclaveRoutes } from "./enclaves/EnclaveRoutes";
import { EnclavesContextProvider } from "./enclaves/EnclavesContext";
import Experiments from "./experiments/Experiments";
import { ExperimentsContextProvider } from "./experiments/ExperimentsContext";
import { Navbar } from "./Navbar";
import { SettingsContextProvider } from "./settings";

const logLogo = (t: string) => console.log(`%c ${t}`, "background: black; color: #00C223");
logLogo(`
                                                ///////////////////
                    //////////                 ///////////////////
                 .////     ,///             /////          ////*
               /////        ///           /////         /////
            ,////        ,////         *////          ////*
             //        /////         /////         /////
                    *////         *////          ////*
                  /////         /////         /////
               *////         /////          /////
             .////         /////         /////
            .///        /////          ////*        //
            ///.      /////         //////          /////
            ////                  ////*.////          *////
             ////              /////      /////          /////
              /////         *////*          .////          *////
                 //////////////                ////////////////////

`);

console.log(`Kurtosis web UI version: ${process.env.REACT_APP_VERSION || "Unknown"}`);

export const EmuiApp = () => {
  return (
    <SettingsContextProvider>
      <KurtosisThemeProvider>
        <KurtosisPackageIndexerProvider>
          <KurtosisClientProvider>
            <KurtosisRouter />
          </KurtosisClientProvider>
        </KurtosisPackageIndexerProvider>
      </KurtosisThemeProvider>
    </SettingsContextProvider>
  );
};

const KurtosisRouter = () => {
  const kurtosisClient = useKurtosisClient();

  const router = useMemo(
    () =>
      createBrowserRouter(
        [
          {
            element: (
              <ExperimentsContextProvider>
                <AppLayout navbar={<Navbar />}>
                  <Outlet />
                </AppLayout>
              </ExperimentsContextProvider>
            ),
            children: [
              {
                path: "/",
                element: (
                  <EnclavesContextProvider>
                    <Outlet />
                    <CreateEnclave />
                    <BuildEnclave />
                  </EnclavesContextProvider>
                ),
                children: enclaveRoutes(),
              },
              {
                path: "/catalog",
                element: (
                  <CatalogContextProvider>
                    <Outlet />
                  </CatalogContextProvider>
                ),
                children: catalogRoutes(),
              },
              {
                path: "/experiments",
                handle: {
                  type: "experiments" as "experiments",
                  crumb: () => ({ name: "Experiments", destination: "/experiments" }),
                },
                id: "experiments",
                element: <Experiments />,
              },
            ],
          },
        ],
        {
          basename: kurtosisClient.getBaseApplicationUrl().pathname,
        },
      ),
    [kurtosisClient],
  );

  return <RouterProvider router={router} />;
};
