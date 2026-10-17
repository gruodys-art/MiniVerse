import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The Go server embeds cmd/web/dist. Dev mode proxies the API so the
// client can be rebuilt without restarting the simulation.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
  build: {
    outDir: "../cmd/web/dist",
    emptyOutDir: true,
    sourcemap: true,
  },
});
