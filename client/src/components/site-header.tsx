import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { useLocation } from "react-router-dom";
import { NavActions } from "./nav-actions";
import Notifications from "./notifications";

interface PageInfo {
  title: string;
  description: string;
}

type PageTitlesMap = Record<string, PageInfo>;

export function SiteHeader() {
  const location = useLocation();

  const pageTitles: PageTitlesMap = {
    "/": {
      title: "Home",
      description: "Overview of system stats and activity"
    },
    "/home": {
      title: "Home",
      description: "Overview of system stats and activity"
    },
    "/logs": {
      title: "Logs",
      description: "View real-time and historical DNS logs"
    },
    "/domains": {
      title: "Domains",
      description: "Manage monitored or filtered domains"
    },
    "/blacklist": {
      title: "Blacklist",
      description: "Block specific domains from resolving"
    },
    "/whitelist": {
      title: "Whitelist",
      description: "Allow specific domains to bypass filters"
    },
    "/resolution": {
      title: "Resolution",
      description: "Configure custom DNS entries"
    },
    "/prefetch": {
      title: "Prefetch",
      description: "Manage DNS prefetching settings"
    },
    "/upstream": {
      title: "Upstream",
      description: "Configure upstream DNS servers"
    },
    "/clients": {
      title: "Clients",
      description: "See connected clients and their activity"
    },
    "/settings": {
      title: "Settings",
      description: "Customize server behavior and UI options"
    },
    "/changelog": {
      title: "Changelog",
      description: "View recent changes and release notes"
    }
  };

  const currentPage = pageTitles[location.pathname];
  const title = currentPage?.title || "Unknown Page";
  const description = currentPage?.description || "";

  return (
    <header className="flex h-12 shrink-0 items-center gap-2 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
      <div className="flex w-full items-center gap-2 px-4 lg:gap-3 lg:px-6">
        <SidebarTrigger className="-ml-1 hover:bg-accent hover:text-accent-foreground" />

        <Separator
          orientation="vertical"
          className="mx-1 h-4 data-[orientation=vertical]:h-4"
        />

        <div className="flex items-center gap-3 min-w-0 flex-1">
          <h1 className="text-base font-semibold tracking-tight text-foreground">
            {title}
          </h1>

          {description && (
            <div className="hidden sm:block">
              <span className="inline-flex items-center rounded-md bg-muted px-2 py-1 text-xs font-medium text-muted-foreground ring-1 ring-inset ring-border">
                {description}
              </span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 px-4 lg:px-6">
        <Notifications />
        <NavActions />
      </div>
    </header>
  );
}
