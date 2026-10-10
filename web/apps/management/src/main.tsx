import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AppTheme } from "shared/theme";
import "shared/styles.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { createManagementRouter } from "./router";

const queryClient = new QueryClient();
const router = createManagementRouter(queryClient);

const root = document.getElementById("root");
if (!root) throw new Error("Missing root element");

createRoot(root).render(
  <StrictMode>
    <AppTheme>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AppTheme>
  </StrictMode>,
);
