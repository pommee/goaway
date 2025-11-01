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
import {
  BrowserIcon,
  CloudArrowUpIcon,
  GearIcon,
  GithubLogoIcon,
  HouseIcon,
  ListIcon,
  NoteIcon,
  NotebookIcon,
  PersonSimpleThrowIcon,
  SignOutIcon,
  TrafficSignIcon,
  UsersIcon
} from "@phosphor-icons/react";
import * as React from "react";
import { TextAnimate } from "./ui/text-animate";
import { ServerStatistics } from "./server-statistics";

const data = {
  navMain: [
    {
      title: "Home",
      url: "/home",
      icon: HouseIcon
    },
    {
      title: "Logs",
      url: "/logs",
      icon: NotebookIcon
    },
    {
      title: "Lists",
      url: "/blacklist",
      icon: ListIcon,
      items: [
        {
          title: "Blacklist",
          url: "/blacklist"
        },
        {
          title: "Whitelist",
          url: "/whitelist"
        }
      ]
    },
    {
      title: "Resolution",
      url: "/resolution",
      icon: TrafficSignIcon
    },
    {
      title: "Prefetch",
      url: "/prefetch",
      icon: PersonSimpleThrowIcon
    },
    {
      title: "Upstream",
      url: "/upstream",
      icon: CloudArrowUpIcon
    },
    {
      title: "Clients",
      url: "/clients",
      icon: UsersIcon
    },
    {
      title: "Settings",
      url: "/settings",
      icon: GearIcon
    },
    {
      title: "Changelog",
      url: "/changelog",
      icon: NoteIcon
    }
  ],
  navSecondary: [
    {
      title: "Website",
      url: "https://github.com/pommee/goaway",
      icon: BrowserIcon,
      blank: "_blank"
    },
    {
      title: "GitHub",
      url: "https://github.com/pommee/goaway",
      icon: GithubLogoIcon,
      blank: "_blank"
    },
    {
      title: "Logout",
      url: "/login",
      icon: SignOutIcon,
      blank: ""
    }
  ]
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <div className="border-r border-accent">
      <Sidebar variant="inset" {...props}>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <a href="/home">
                  <img src={"/logo.png"} alt={"project-mascot"} width={50} />
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
        <ServerStatistics />
        <SidebarContent>
          <NavMain items={data.navMain} />
          <NavSecondary items={data.navSecondary} className="mt-auto" />
        </SidebarContent>
      </Sidebar>
    </div>
  );
}
