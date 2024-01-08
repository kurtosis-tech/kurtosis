import { AppLayout, KurtosisThemeProvider } from "kurtosis-ui-components";
import { useMemo } from "react";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { KurtosisClientProvider, useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import { KurtosisPackageIndexerProvider } from "../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { CatalogContextProvider } from "./catalog/CatalogContext";
import { catalogRoutes } from "./catalog/CatalogRoutes";
import { CreateEnclave } from "./enclaves/components/CreateEnclave";
import { enclaveRoutes } from "./enclaves/EnclaveRoutes";
import { EnclavesContextProvider } from "./enclaves/EnclavesContext";
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
          <KurtosisRouter />
        </KurtosisClientProvider>
      </KurtosisPackageIndexerProvider>
    </KurtosisThemeProvider>
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
              <AppLayout navbar={<Navbar />}>
                <Outlet />
              </AppLayout>
            ),
            children: [
              {
                path: "/",
                element: (
                  <EnclavesContextProvider>
                    <Outlet />
                    <CreateEnclave />
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
