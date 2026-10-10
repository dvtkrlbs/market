import { defineConfig } from "vite-plus";

export default defineConfig({
  fmt: {},
  lint: {
    jsPlugins: [{ name: "vite-plus", specifier: "vite-plus/oxlint-plugin" }],
    rules: { "vite-plus/prefer-vite-plus-imports": "error" },
    options: { typeAware: true, typeCheck: true },
  },
  test: {
    include: ["packages/*/src/**/*.test.ts"],
  },
  run: {
    tasks: {
      openapi: {
        command: "go run ./cmd/openapi-stitch",
        cwd: "..",
        // The Go contract lives outside the JavaScript workspace.
        cache: false,
      },
    },
  },
});
