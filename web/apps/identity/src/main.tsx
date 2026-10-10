import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AppTheme } from "theme";
import "theme/styles.css";
import App from "./App";

const root = document.getElementById("root");
if (!root) throw new Error("Missing root element");

createRoot(root).render(
  <StrictMode>
    <AppTheme>
      <App />
    </AppTheme>
  </StrictMode>,
);
