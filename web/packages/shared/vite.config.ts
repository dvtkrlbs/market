import { defineConfig } from "vite-plus";

export default defineConfig({
  pack: {
    entry: ["src/index.ts", "src/api.ts", "src/theme.tsx"],
    format: ["esm"],
    dts: true,
    sourcemap: true,
  },
  run: {
    tasks: {
      generate: {
        command: "orval --config orval.config.ts",
        dependsOn: ["market-web#openapi"],
        cache: false,
      },
      build: {
        command: "vp pack",
        dependsOn: ["generate"],
      },
    },
  },
});
