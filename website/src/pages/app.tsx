import { ThemeProvider } from "@/components/theme-provider";
import Layout from "../app/layout";
import { Routes, Route } from "react-router-dom";
import Logs from "@/pages/logs";
import Home from "@/pages/home";
import Settings from "./settings";
import { Toaster } from "@/components/ui/sonner";

function App() {
  return (
    <Layout>
      <div className="h-full w-full bg-sidebar rounded-sm border-1 border-accent p-2">
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
          <Routes>
            <Route path="/home" element={<Home />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </ThemeProvider>
      </div>
      <Toaster />
    </Layout>
  );
}

export default App;
