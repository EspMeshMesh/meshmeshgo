import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [react()],
  server: {
    host: true,
    proxy: {
      "/api/v1": {
        target: "http://localhost:4040",
        changeOrigin: true,
        //rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
    },
  },
  build: {
    outDir: "../managerui/data",
    emptyOutDir: true,
    sourcemap: mode === "development",
  },
  base: "./",
}));
