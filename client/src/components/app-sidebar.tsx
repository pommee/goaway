import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem
} from "@/components/ui/sidebar";
import { GenerateQuote } from "@/quotes";
import { SiGithub } from "@icons-pack/react-simple-icons";
import {
  CloudArrowUp,
  ComputerTower,
  Gear,
  House,
  List,
  Note,
  Notebook,
  SignOut,
  TrafficSign,
  Users
} from "@phosphor-icons/react";
import * as React from "react";
import { ServerStatistics } from "./server-statistics";
import { Separator } from "./ui/separator";
import { TextAnimate } from "./ui/text-animate";

const data = {
  navMain: [
    {
      title: "Home",
      url: "/home",
      icon: House
    },
    {
      title: "Logs",
      url: "/logs",
      icon: Notebook
    },
    //    {
    //      title: "Domains",
    //      url: "/domains",
    //      icon: Server,
    //    },
    {
      title: "Lists",
      url: "/lists",
      icon: List
    },
    {
      title: "Resolution",
      url: "/resolution",
      icon: TrafficSign
    },
    {
      title: "Upstream",
      url: "/upstream",
      icon: CloudArrowUp
    },
    {
      title: "Clients",
      url: "/clients",
      icon: Users
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Gear
    },
    {
      title: "Changelog",
      url: "/changelog",
      icon: Note
    },
    {
      title: "System",
      url: "/system",
      icon: ComputerTower
    }
  ],
  navSecondary: [
    {
      title: "GitHub",
      url: "https://github.com/pommee/goaway",
      icon: SiGithub,
      blank: "_blank"
    },
    {
      title: "Logout",
      url: "/login",
      icon: SignOut,
      blank: ""
    }
  ]
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <div className="border-r-1 border-accent">
      <Sidebar variant="inset" {...props}>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <a href="/home">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg">
                    <img src={"/logo.png"} />
                  </div>
                  <div className="grid flex-1 text-left text-lg leading-tight">
                    <span className="truncate font-medium">GoAway</span>
                    <TextAnimate
                      className="truncate text-xs"
                      animation="blurInUp"
                      by="character"
                      once
                    >
                      {GenerateQuote()}
                    </TextAnimate>
                    <span></span>
                  </div>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <Separator />
        <ServerStatistics />
        <SidebarContent>
          <NavMain items={data.navMain} />
          <NavSecondary items={data.navSecondary} className="mt-auto" />
        </SidebarContent>
      </Sidebar>
    </div>
  );
}
