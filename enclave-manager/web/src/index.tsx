import React from "react";
import ReactDOM from "react-dom/client";
import { EmuiApp } from "./emui/App";

const root = ReactDOM.createRoot(document.getElementById("root") as HTMLElement);
root.render(
  <React.StrictMode>
    <EmuiApp />
  </React.StrictMode>,
);
