import * as React from "react";
import {
  ArrowUpLeftSquare,
  ChartNoAxesGantt,
  Command,
  GithubIcon,
  Home,
  ListFilter,
  LogOut,
  Logs,
  Server,
  Settings,
  User,
} from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { ServerStatistics } from "./server-statistics";

const data = {
  user: {
    name: "admin",
    email: "admin@user.com",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain: [
    {
      title: "Home",
      url: "/home",
      icon: Home,
    },
    {
      title: "Logs",
      url: "/logs",
      icon: Logs,
    },
    {
      title: "Domains",
      url: "/domains",
      icon: Server,
    },
    {
      title: "Lists",
      url: "/lists",
      icon: ListFilter,
    },
    {
      title: "Upstream",
      url: "/upstream",
      icon: ArrowUpLeftSquare,
    },
    {
      title: "Clients",
      url: "/clients",
      icon: User,
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Settings,
    },
    {
      title: "Changelog",
      url: "/changelog",
      icon: ChartNoAxesGantt,
    },
  ],
  navSecondary: [
    {
      title: "GitHub",
      url: "https://github.com/pommee/goaway",
      icon: GithubIcon,
      blank: "_blank",
    },
    {
      title: "Logout",
      url: "#",
      icon: LogOut,
      blank: "",
    },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <div className="border-r-1 border-accent">
      <Sidebar variant="inset" {...props}>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <a href="#">
                  <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                    <Command className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-lg leading-tight">
                    <span className="truncate font-medium">GoAway</span>
                    <span className="truncate text-xs">Block ads!</span>
                  </div>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <ServerStatistics />
        <SidebarContent>
          <NavMain items={data.navMain} />
          <NavSecondary items={data.navSecondary} className="mt-auto" />
        </SidebarContent>
      </Sidebar>
    </div>
  );
}
