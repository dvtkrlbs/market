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
      baseUrl: {
        runtime: "getApiBaseUrl()",
        imports: [{ name: "getApiBaseUrl", importPath: "../config" }],
      },
      override: {
        requestOptions: { credentials: "include" },
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
