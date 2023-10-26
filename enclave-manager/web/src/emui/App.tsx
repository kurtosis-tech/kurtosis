import { Box } from "@chakra-ui/react";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";
import { KurtosisClientProvider } from "../client/KurtosisClientContext";
import { AppLayout } from "../components/AppLayout";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";
import { enclaveRoutes } from "./enclaves/Enclaves";
import { Navbar } from "./Navbar";

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
