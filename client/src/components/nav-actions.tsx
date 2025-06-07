"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
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
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import {
  ClockIcon,
  CloudArrowUpIcon,
  DotsThreeOutlineIcon,
  InfoIcon,
  PauseIcon,
  PlayCircleIcon
} from "@phosphor-icons/react";
import { JSX, useEffect, useState } from "react";
import { toast } from "sonner";
import { Metrics } from "./server-statistics";
import { Input } from "./ui/input";
import { compare } from "compare-versions";

const data = [
  [
    {
      label: "About",
      icon: InfoIcon,
      dialog: AboutDialog
    },
    {
      label: "Check for update",
      icon: CloudArrowUpIcon,
      dialog: UpdateDialog
    }
  ],
  [
    {
      label: "Blocking",
      icon: PauseIcon,
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
              {(responseData?.commit && (
                <a
                  href={
                    "https://github.com/pommee/goaway/commit/" +
                    responseData?.commit
                  }
                  target="_blank"
                >
                  {responseData?.commit.substring(0, 8)}
                </a>
              )) ||
                "Not available"}
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
  useEffect(() => {
    const installedVersion = localStorage.getItem("installedVersion");

    async function lookForUpdate() {
      try {
        localStorage.setItem("lastUpdateCheck", Date.now().toString());
        const response = await fetch(
          "https://api.github.com/repos/pommee/goaway/tags"
        );
        const data = await response.json();
        const latestVersion = data[0].name.replace("v", "");
        localStorage.setItem("latestVersion", latestVersion);

        if (compare(latestVersion, installedVersion, "<=")) {
          toast.info("No new version found!");
        }
      } catch (error) {
        console.error("Failed to check for updates:", error);
        return null;
      }
    }

    lookForUpdate();
  });
}

export default function PauseBlockingDialog() {
  type PausedResponse = {
    paused: boolean;
    timeLeft: number;
  };

  const [time, setTime] = useState("10");
  const [isLoading, setIsLoading] = useState(false);
  const [pauseStatus, setPauseStatus] = useState<PausedResponse>();
  const [remainingTime, setRemainingTime] = useState(0);

  useEffect(() => {
    const fetchPauseStatus = async () => {
      try {
        const [status, response] = await GetRequest("pause");
        if (status === 200) {
          setPauseStatus(response);

          if (response.paused) {
            setRemainingTime(response.timeLeft);
          }
        }
      } catch (error) {
        console.error("Error fetching pause status:", error);
      }
    };

    fetchPauseStatus();

    const intervalId = setInterval(() => {
      if (pauseStatus?.paused) {
        if (remainingTime > 0) {
          setRemainingTime((prevTime) => Math.max(0, prevTime - 1));
        } else {
          fetchPauseStatus();
        }
      }
    }, 1000);

    return () => clearInterval(intervalId);
  }, [pauseStatus?.paused, remainingTime]);

  const handlePause = async () => {
    setIsLoading(true);
    try {
      const [status] = await PostRequest("pause", {
        time: parseInt(time)
      });

      if (status === 200) {
        toast.info(`Paused blocking for ${time} seconds`);
        const [getStatus, getResponse] = await GetRequest("pause");
        if (getStatus === 200) {
          setPauseStatus(getResponse);
          if (getResponse.paused) {
            setRemainingTime(getResponse.timeLeft);
          }
        }
      } else {
        console.error("Failed to pause blocking");
        toast.error("Failed to pause blocking");
      }
    } catch (error) {
      console.error("Error pausing blocking:", error);
      toast.error("Error pausing blocking");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemovePause = async () => {
    setIsLoading(true);
    try {
      const [status] = await DeleteRequest("pause", null);

      if (status === 200) {
        toast.success("Blocking resumed");
        setPauseStatus((prev) => ({ ...prev, paused: false }));
        setRemainingTime(0);
      } else {
        console.error("Failed to resume blocking");
        toast.error("Failed to resume blocking");
      }
    } catch (error) {
      console.error("Error resuming blocking:", error);
      toast.error("Error resuming blocking");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <DialogContent className="sm:max-w-md">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <ClockIcon className="text-blue-400" />
          {pauseStatus?.paused ? "Blocking Paused" : "Pause Blocking"}
        </DialogTitle>
        <DialogDescription className="text-sm text-gray-500">
          {pauseStatus?.paused
            ? `Blocking is currently paused. Remaining time: ${remainingTime} seconds.`
            : "This will temporarily pause domain blocking, allowing all traffic to pass through."}
        </DialogDescription>
      </DialogHeader>

      {!pauseStatus?.paused ? (
        <>
          <div className="py-4">
            <label
              htmlFor="pause-time"
              className="block text-sm font-medium mb-2"
            >
              Duration (seconds)
            </label>
            <Input
              id="pause-time"
              type="number"
              min="1"
              value={time}
              onChange={(e) => setTime(e.target.value)}
              className="w-full"
            />
          </div>

          <DialogFooter className="flex justify-end gap-2">
            <Button variant="outline" className="border-gray-300">
              Cancel
            </Button>
            <Button
              onClick={handlePause}
              disabled={isLoading}
              className="bg-blue-500 hover:bg-blue-600 text-white"
            >
              {isLoading ? "Pausing..." : "Pause Blocking"}
            </Button>
          </DialogFooter>
        </>
      ) : (
        <DialogFooter className="flex justify-center mt-4">
          <Button
            onClick={handleRemovePause}
            disabled={isLoading}
            className="bg-green-500 hover:bg-green-600 text-white flex items-center gap-2"
          >
            <PlayCircleIcon size={18} />
            {isLoading ? "Resuming..." : "Resume Blocking Now"}
          </Button>
        </DialogFooter>
      )}
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
            <DotsThreeOutlineIcon />
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
