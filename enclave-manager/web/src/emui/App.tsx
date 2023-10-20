import React from "react";
import { Box } from "@chakra-ui/react";
import { AppLayout } from "../components/AppLayout";
import { Navbar } from "./Navbar";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { enclaveRoutes } from "./enclaves/Enclaves";
import { KurtosisClientProvider } from "../client/KurtosisClientContext";

const router = createBrowserRouter([
  {
    element: (
      <AppLayout Nav={<Navbar />}>
        <KurtosisClientProvider>
          <Outlet />
        </KurtosisClientProvider>
      </AppLayout>
    ),
    children: [
      { path: "/", children: enclaveRoutes },
      { path: "/catalog", element: <Box>Goodby World</Box> },
    ],
  },
]);

export const EmuiApp = () => {
  return (
    <KurtosisThemeProvider>
      <RouterProvider router={router} />
    </KurtosisThemeProvider>
  );
};
