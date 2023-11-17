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
import { LocationBroadcaster } from "../components/LocationBroadcaster";
import { LocationListener } from "../components/LocationListener";
import { catalogRoutes } from "./catalog/CatalogRoutes";
import { EmuiAppContextProvider } from "./EmuiAppContext";
import { enclaveRoutes } from "./enclaves/EnclaveRoutes";
import { Navbar } from "./Navbar";

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
              <AppLayout Nav={<Navbar baseApplicationUrl={kurtosisClient.getBaseApplicationUrl()} />}>
                <Outlet />
                <CreateEnclave />
                <LocationBroadcaster />
                <LocationListener />
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
