import { hello } from "kurtosis-ui-components";
import React from "react";
import ReactDOM from "react-dom/client";
import { EmuiApp } from "./emui/App";

hello();

const root = ReactDOM.createRoot(document.getElementById("root") as HTMLElement);
root.render(
  <React.StrictMode>
    <EmuiApp />
  </React.StrictMode>,
);
