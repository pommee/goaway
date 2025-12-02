import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: ["babel-plugin-react-compiler"]
      }
    }),
    tailwindcss()
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src")
    }
  },
  build: {
    assetsDir: "assets",
    sourcemap: false,
    cssCodeSplit: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules")) {
            if (id.includes("react") || id.includes("react-dom"))
              return "react-vendor";
            if (id.includes("motion")) return "motion";
            if (
              id.includes("lucide-react") ||
              id.includes("@phosphor-icons/react")
            )
              return "icons";
            return "vendor";
          }
          if (id.includes("/src/pages/")) return "pages";
          if (id.includes("/src/components/")) return "components";
          return null;
        }
      }
    }
  }
});
