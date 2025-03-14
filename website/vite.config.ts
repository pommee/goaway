import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [
    react(), 
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    assetsDir: "assets",
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom')) {
              return 'react-vendor'
            }
            if (id.includes('tailwindcss') || id.includes('@tailwindcss')) {
              return 'tailwind-vendor'
            }
            return 'vendor'
          }
          
          if (id.includes('/src/components/') || id.includes('/src/pages/')) {
            return 'app-chunks'
          }
        }
      }
    }
  }
});
