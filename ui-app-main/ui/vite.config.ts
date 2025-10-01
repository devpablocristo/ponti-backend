import { defineConfig, ConfigEnv } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig(({ mode }: ConfigEnv) => {
  return {
    plugins: [react()],
    base: "/",
    server:
      mode === "development"
        ? {
            proxy: {
              "/api": {
                target: "http://localhost:3000/api",
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, ""),
              },
            },
          }
        : undefined,
    // resolve: {
    //   alias: {
    //     "@": path.resolve(__dirname, "./src"),
    //   },
    // },
  };
});
