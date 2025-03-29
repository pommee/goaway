import { lazy, Suspense } from "react";
import { ThemeProvider } from "@/components/theme-provider";
import Layout from "../app/layout";
import { Routes, Route } from "react-router-dom";
import { Toaster } from "@/components/ui/sonner";

const Home = lazy(() => import("@/pages/home"));
const Settings = lazy(() => import("./settings"));
const Changelog = lazy(() => import("./changelog"));
const Clients = lazy(() => import("./clients"));
const Lists = lazy(() => import("./lists"));
const Logs = lazy(() => import("./logs"));
const Login = lazy(() => import("./login"));

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <Suspense>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={<Layout />}>
            <Route path="/" element={<Home />} />
            <Route path="/home" element={<Home />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="/lists" element={<Lists />} />
            <Route path="/clients" element={<Clients />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/changelog" element={<Changelog />} />
          </Route>
        </Routes>
      </Suspense>
      <Toaster />
    </ThemeProvider>
  );
}

export default App;
