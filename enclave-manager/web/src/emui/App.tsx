import React from "react";
import { Box } from "@chakra-ui/react";
import { AppLayout } from "../components/AppLayout";
import { Navbar } from "./Navbar";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { enclaveRoutes } from "./enclaves/Enclaves";

const router = createBrowserRouter([
  {
    element: (
      <AppLayout Nav={<Navbar />}>
        <Outlet />
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
