import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { useLocation } from "react-router-dom";
import { NavActions } from "./nav-actions";
import Notifications from "./notifications";

export function SiteHeader() {
  const location = useLocation();

  const pageTitles: { [key: string]: string } = {
    "/": "Home",
    "/home": "Home",
    "/logs": "Logs",
    "/domains": "Domains",
    "/lists": "Lists",
    "/resolution": "Resolution",
    "/prefetch": "Resolution",
    "/upstream": "Upstream",
    "/clients": "Clients",
    "/settings": "Settings",
    "/changelog": "Changelog"
  };

  const title = pageTitles[location.pathname] || "";

  return (
    <header className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <h1 className="text-base font-medium">{title}</h1>
      </div>
      <Notifications />
      <NavActions />
    </header>
  );
}
