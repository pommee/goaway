import * as React from "react";
import { SiGithub, SiWire } from "@icons-pack/react-simple-icons";
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
import { ServerStatistics } from "./server-statistics";
import { Separator } from "./ui/separator";
import { TextAnimate } from "./ui/text-animate";
import {
  Gear,
  House,
  List,
  Note,
  Notebook,
  SignOut,
  Users
} from "@phosphor-icons/react";

const quotes = [
  "Block party!",
  "No ads!",
  "Bye-bye, spam!",
  "Adios, ads!",
  "Get lost!",
  "Bye, trackers!",
  "Stop, right there!",
  "Catch you later!",
  "Ad free zone!",
  "Blockzilla strikes!",
  "Ad-ocalypse now!",
  "Nope, not today!",
  "Buzz off, ads!",
  "Ads? Not here!",
  "Shh... no ads.",
  "Ad blocker engaged!",
  "Gone in a click!",
  "Ads begone!",
  "Block mode: ON!",
  "Spam, who?",
  "No entry for ads!",
  "Bye-bye bandwidth hogs!",
  "Ad-free vibes!",
  "Don't block me!",
  "Not in my house!",
  "Get out, ads!",
  "Bye-bye popups!",
  "Say no to ads!",
  "Ad block, power!",
  "Stay adless!",
  "Access denied!",
  "Tracker blocked!",
  "No tracking zone!",
  "Rejected!",
  "Not welcome here!",
  "Request denied!",
  "Nice try, ads!",
  "Privacy shield up!",
  "Blocked by choice!",
  "Ad-free journey!",
  "Nothing to see!",
  "No soliciting!",
  "Request terminated!",
  "Traffic stopped!",
  "Privacy first!",
  "No way through!",
  "Blocked successfully!",
  "Shields activated!",
  "Look elsewhere!",
  "Access restricted!",
  "Stop tracking me!",
  "Blackholed!",
  "Sinkhole active!",
  "No ads allowed!",
  "Tracker neutralized!",
  "Request sunk!",
  "Bye data miners!",
  "Privacy protected!",
  "Tracking blocked!",
  "Not happening!"
];

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
      icon: SiWire
    },
    // {
    //  title: "Upstream",
    //  url: "/upstream",
    //  icon: ArrowUpLeftSquare,
    //},
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
  const quote = quotes[Math.floor(Math.random() * quotes.length)];

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
                      {quote}
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
