import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { AnimatePresence } from "motion/react";
import { Route, Routes, useLocation } from "react-router-dom";
import Layout from "../app/layout";
import { Blacklist } from "./blacklist";
import Changelog from "./changelog";
import { Clients } from "./clients";
import { Home } from "./home";
import { Logs } from "./logs";
import { Prefetch } from "./prefetch";
import { Resolution } from "./resolution";
import { Settings } from "./settings";
import { Upstream } from "./upstream";
import { Whitelist } from "./whitelist";
import { GenerateQuote } from "@/quotes";
import Login from "./login";
import { FileXIcon } from "@phosphor-icons/react";

function NotFound() {
  return (
    <div className="flex items-center justify-center mt-[20%] px-4">
      <div className="flex flex-col items-center text-center max-w-md">
        <FileXIcon className="w-16 h-16 text-muted-foreground mb-6" />
        <h2 className="text-2xl font-semibold mb-3">Page Not Found</h2>
        <p className="text-muted-foreground">
          The page{" "}
          <span className="inline-block bg-muted-foreground/10 px-2 py-0.5 rounded-sm text-sm">
            {window.location.pathname}
          </span>{" "}
          does not exist.
        </p>
      </div>
    </div>
  );
}

function App() {
  const location = useLocation();

  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AnimatePresence mode="wait">
        <Routes location={location} key={location.pathname}>
          <Route path="/login" element={<Login quote={GenerateQuote()} />} />
          <Route element={<Layout />}>
            <Route path="/" element={<Home />} />
            <Route path="/home" element={<Home />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="/blacklist" element={<Blacklist />} />
            <Route path="/whitelist" element={<Whitelist />} />
            <Route path="/resolution" element={<Resolution />} />
            <Route path="/prefetch" element={<Prefetch />} />
            <Route path="/upstream" element={<Upstream />} />
            <Route path="/clients" element={<Clients />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/changelog" element={<Changelog />} />
            <Route path="*" element={<NotFound />} />
          </Route>
        </Routes>
      </AnimatePresence>
      <Toaster />
    </ThemeProvider>
  );
}

export default App;
