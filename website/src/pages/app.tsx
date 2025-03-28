import { ThemeProvider } from "@/components/theme-provider";
import Layout from "../app/layout";
import { Routes, Route } from "react-router-dom";
import Home from "@/pages/home";
import Settings from "./settings";
import { Toaster } from "@/components/ui/sonner";
import Changelog from "./changelog";
import Clients from "./clients";
import Lists from "./lists";
import Logs from "./logs";
import Login from "./login";

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
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
      <Toaster />
    </ThemeProvider>
  );
}

export default App;
