import { astryxStylex } from "@astryxdesign/build/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite-plus";

export function defineAppConfig(rootDir: string, port: number) {
  return defineConfig(({ command }) => ({
    plugins: [...astryxStylex({ rootDir, dev: command === "serve" }), react()],
    server: { port, strictPort: true },
    preview: { port, strictPort: true },
  }));
}
