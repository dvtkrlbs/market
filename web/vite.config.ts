import { astryxStylex } from "@astryxdesign/build/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite-plus";

// Component tests need StyleX transforms, not the development CSS middleware.
// Its polling timer relies on an HTTP server, which Vitest does not start.
const testStylexPlugins = astryxStylex({ rootDir: import.meta.dirname, dev: false }).map(
  (plugin) => ({ ...plugin, configureServer: undefined }),
);

export default defineConfig({
  plugins: [...testStylexPlugins, react()],
  fmt: { ignorePatterns: ["**/routeTree.gen.ts"] },
  lint: {
    ignorePatterns: ["**/routeTree.gen.ts"],
    jsPlugins: [{ name: "vite-plus", specifier: "vite-plus/oxlint-plugin" }],
    rules: { "vite-plus/prefer-vite-plus-imports": "error" },
    options: { typeAware: true, typeCheck: true },
  },
  test: {
    include: ["packages/*/src/**/*.test.ts", "apps/*/src/**/*.test.tsx"],
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
