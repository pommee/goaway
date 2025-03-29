import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    assetsDir: "assets",
    sourcemap: false,
    cssCodeSplit: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules")) {
            if (id.includes("react") || id.includes("react-dom")) {
              return "react-core";
            }
            if (id.includes("react-router-dom")) {
              return "router";
            }

            if (id.includes("@radix-ui") || id.includes("@shadcn")) {
              return "ui-vendor";
            }
            if (id.includes("lucide-react")) {
              return "icons";
            }
            if (id.includes("recharts")) {
              return "charts";
            }

            if (id.includes("motion") || id.includes("tw-animate-css")) {
              return "animations";
            }

            if (
              id.includes("tailwind") ||
              id.includes("clsx") ||
              id.includes("tailwind-merge")
            ) {
              return "tailwind-utils";
            }

            if (
              id.includes("class-variance-authority") ||
              id.includes("sonner")
            ) {
              return "utils";
            }

            return "vendor";
          }

          if (id.includes("/src/components/ui/")) {
            return "ui-components";
          }

          if (id.includes("/src/components/")) {
            return "components";
          }

          if (id.includes("/src/pages/")) {
            return "pages";
          }

          if (id.includes("/src/hooks/") || id.includes("/src/utils/")) {
            return "shared";
          }
        },
        chunkFileNames: "assets/[name]-[hash].js",
        entryFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
  },
});
