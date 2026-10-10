import { astryxStylex } from "@astryxdesign/build/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, type PluginOption } from "vite-plus";

export function defineAppConfig(rootDir: string, port: number, plugins: PluginOption[] = []) {
  return defineConfig(({ command }) => ({
    plugins: [...plugins, ...astryxStylex({ rootDir, dev: command === "serve" }), react()],
    envDir: import.meta.dirname,
    server: { port, strictPort: true },
    preview: { port, strictPort: true },
  }));
}
