import React from "react";
import { Box } from "@chakra-ui/react";
import { AppLayout } from "../components/AppLayout";
import { Navbar } from "./Navbar";
import { KurtosisThemeProvider } from "../components/KurtosisThemeProvider";

export const EmuiApp = () => {
  return (
    <KurtosisThemeProvider>
      <AppLayout Nav={<Navbar />}>
        <Box>Hello World</Box>
      </AppLayout>
    </KurtosisThemeProvider>
  );
};
