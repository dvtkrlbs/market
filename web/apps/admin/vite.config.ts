import { tanstackRouter } from "@tanstack/router-plugin/vite";
import { defineAppConfig } from "../../vite.app.config";

export default defineAppConfig(import.meta.dirname, 5174, [
  tanstackRouter({
    target: "react",
    autoCodeSplitting: true,
    quoteStyle: "double",
    semicolons: true,
  }),
]);
