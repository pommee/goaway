import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { AnimatePresence, motion } from "motion/react";
import { Suspense } from "react";
import { Route, Routes, useLocation } from "react-router-dom";
import Layout from "../app/layout";
import { Blacklist } from "./blacklist";
import Changelog from "./changelog";
import { Clients } from "./clients";
import { Home } from "./home";
import { Login } from "./login";
import { Logs } from "./logs";
import { Prefetch } from "./prefetch";
import { Resolution } from "./resolution";
import { Settings } from "./settings";
import { Upstream } from "./upstream";
import { Whitelist } from "./whitelist";

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-full w-full">
      <div className="flex flex-col items-center">
        <div className="w-16 h-16 border-4 border-t-blue-500 border-r-transparent border-b-blue-500 border-l-transparent rounded-full animate-spin"></div>
        <p className="mt-4 text-gray-400">Loading page...</p>
      </div>
    </div>
  );
}

function PageTransition({ children }) {
  return (
    <motion.div initial="initial" animate="in" exit="out" className="h-full">
      <Suspense fallback={<PageLoader />}>{children}</Suspense>
    </motion.div>
  );
}

function App() {
  const location = useLocation();

  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AnimatePresence mode="wait">
        <Routes location={location} key={location.pathname}>
          <Route
            path="/login"
            element={
              <PageTransition>
                <Login />
              </PageTransition>
            }
          />
          <Route element={<Layout />}>
            <Route
              path="/"
              element={
                <PageTransition>
                  <Home />
                </PageTransition>
              }
            />
            <Route
              path="/home"
              element={
                <PageTransition>
                  <Home />
                </PageTransition>
              }
            />
            <Route
              path="/logs"
              element={
                <PageTransition>
                  <Logs />
                </PageTransition>
              }
            />
            <Route
              path="/blacklist"
              element={
                <PageTransition>
                  <Blacklist />
                </PageTransition>
              }
            />
            <Route
              path="/whitelist"
              element={
                <PageTransition>
                  <Whitelist />
                </PageTransition>
              }
            />
            <Route
              path="/resolution"
              element={
                <PageTransition>
                  <Resolution />
                </PageTransition>
              }
            />
            <Route
              path="/prefetch"
              element={
                <PageTransition>
                  <Prefetch />
                </PageTransition>
              }
            />
            <Route
              path="/upstream"
              element={
                <PageTransition>
                  <Upstream />
                </PageTransition>
              }
            />
            <Route
              path="/clients"
              element={
                <PageTransition>
                  <Clients />
                </PageTransition>
              }
            />
            <Route
              path="/settings"
              element={
                <PageTransition>
                  <Settings />
                </PageTransition>
              }
            />
            <Route
              path="/changelog"
              element={
                <PageTransition>
                  <Changelog />
                </PageTransition>
              }
            />
          </Route>
        </Routes>
      </AnimatePresence>
      <Toaster />
    </ThemeProvider>
  );
}

export default App;
