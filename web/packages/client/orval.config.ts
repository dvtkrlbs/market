import { defineConfig } from "orval";

export default defineConfig({
  market: {
    input: {
      target: "../../../dist/openapi.yaml",
    },
    output: {
      target: "./src/generated/api.ts",
      client: "fetch",
      mode: "single",
      clean: true,
      override: {
        fetch: {
          includeHttpResponseReturnType: true,
        },
      },
    },
    hooks: {
      afterAllFilesWrite: "vp fmt",
    },
  },
});
