"use client";

import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger
} from "@/components/ui/popover";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem
} from "@/components/ui/sidebar";
import { DotsThreeOutline, Info, Pause, Rss } from "@phosphor-icons/react";
import { JSX, useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { GetRequest } from "@/util";
import { Metrics } from "./server-statistics";

const data = [
  [
    {
      label: "About",
      icon: Info,
      dialog: AboutDialog
    },
    {
      label: "[WIP] Check for update",
      icon: Rss,
      dialog: UpdateDialog
    }
  ],
  [
    {
      label: "[WIP] Pause blocking",
      icon: Pause,
      dialog: PauseBlockingDialog
    }
  ]
];

function AboutDialog() {
  const [responseData, setResponseData] = useState<Metrics>();

  useEffect(() => {
    async function fetchData() {
      try {
        const [, data] = await GetRequest("server");
        setResponseData(data);
      } catch {
        return;
      }
    }

    fetchData();
  }, []);

  return (
    <DialogContent className="w-fit">
      <DialogHeader>
        <DialogTitle>About</DialogTitle>
        <DialogDescription />
        <div className="mt-2 text-sm text-muted-foreground">
          <div className="grid grid-cols-[auto_1fr] gap-y-1 items-center">
            <span className="pr-2">Version:</span>
            <span className="text-white">
              {responseData?.version || "Not available"}
            </span>

            <span className="pr-2">Commit:</span>
            <span className="text-white">
              {responseData?.commit || "Not available"}
            </span>

            <span className="pr-2">Date:</span>
            <span className="text-white">
              {responseData?.date || "Not available"}
            </span>
          </div>
        </div>
      </DialogHeader>
    </DialogContent>
  );
}

function UpdateDialog() {
  return (
    <DialogContent className="w-1/2">
      <DialogHeader>
        <DialogTitle>Check for Updates</DialogTitle>
        <DialogDescription>Checking for a new version...</DialogDescription>
      </DialogHeader>
    </DialogContent>
  );
}

function PauseBlockingDialog() {
  return (
    <DialogContent className="w-1/2">
      <DialogHeader>
        <DialogTitle>Pause Blocking</DialogTitle>
        <DialogDescription>
          Are you sure you want to pause blocking? You will be exposed!
        </DialogDescription>
      </DialogHeader>
    </DialogContent>
  );
}

export function NavActions() {
  const [isOpen, setIsOpen] = useState(false);
  const [DialogComponent, setDialogComponent] = useState<
    null | (() => JSX.Element)
  >(null);

  return (
    <div className="text-sm mr-4">
      <Popover open={isOpen} onOpenChange={setIsOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 data-[state=open]:bg-accent"
          >
            <DotsThreeOutline />
          </Button>
        </PopoverTrigger>
        <PopoverContent
          className="w-56 overflow-hidden rounded-lg p-0"
          align="end"
        >
          <Sidebar collapsible="none" className="bg-transparent">
            <SidebarContent>
              {data.map((group, index) => (
                <SidebarGroup key={index} className="border-b last:border-none">
                  <SidebarGroupContent className="gap-0">
                    <SidebarMenu>
                      {group.map((item, index) => (
                        <SidebarMenuItem key={index}>
                          <SidebarMenuButton
                            onClick={() => {
                              setIsOpen(false);
                              setDialogComponent(() => item.dialog);
                            }}
                          >
                            <item.icon /> <span>{item.label}</span>
                          </SidebarMenuButton>
                        </SidebarMenuItem>
                      ))}
                    </SidebarMenu>
                  </SidebarGroupContent>
                </SidebarGroup>
              ))}
            </SidebarContent>
          </Sidebar>
        </PopoverContent>
      </Popover>

      {DialogComponent && (
        <Dialog
          open={!!DialogComponent}
          onOpenChange={(open) => {
            if (!open) setDialogComponent(null);
          }}
        >
          <DialogComponent />
        </Dialog>
      )}
    </div>
  );
}
