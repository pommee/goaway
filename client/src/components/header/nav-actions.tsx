"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
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
  ArrowsClockwiseIcon,
  ClockIcon,
  CloudArrowUpIcon,
  DotsThreeOutlineIcon,
  InfoIcon,
  PauseIcon,
  PlayCircleIcon,
  WarningIcon
} from "@phosphor-icons/react";
import { compare } from "compare-versions";
import { JSX, useEffect, useState } from "react";
import { toast } from "sonner";
import { Metrics } from "../server-statistics";
import { Input } from "../ui/input";
import { ToggleGroup, ToggleGroupItem } from "../ui/toggle-group";

const data = [
  [
    {
      label: "About",
      icon: InfoIcon,
      dialog: AboutDialog,
      color: "text-blue-600"
    },
    {
      label: "Check for update",
      icon: CloudArrowUpIcon,
      dialog: CheckForUpdate,
      color: "text-yellow-600"
    },
    {
      label: "Restart",
      icon: ArrowsClockwiseIcon,
      dialog: Restart,
      color: "text-yellow-600"
    }
  ],
  [
    {
      label: "Blocking",
      icon: PauseIcon,
      dialog: PauseBlockingDialog,
      color: "text-red-600"
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
    <DialogContent className="lg:w-fit">
      <DialogHeader>
        <DialogTitle className="flex">
          <InfoIcon className="mr-2 text-blue-500" /> About
        </DialogTitle>
        <DialogDescription />
        <div className="mt-2 text-sm">
          <div className="grid grid-cols-[auto_1fr] gap-y-1 items-center">
            <span className="pr-2 text-muted-foreground">Version:</span>
            <span>{responseData?.version || "Not available"}</span>

            <span className="pr-2 text-muted-foreground">Commit:</span>
            <span className="text-blue-400 underline cursor-pointer text-ellipsis overflow-x-hidden">
              {(responseData?.commit && (
                <a
                  href={
                    "https://github.com/pommee/goaway/commit/" +
                    responseData?.commit
                  }
                  target="_blank"
                >
                  {responseData?.commit}
                </a>
              )) ||
                "Not available"}
            </span>

            <span className="pr-2 text-muted-foreground">Date:</span>
            <span>{responseData?.date || "Not available"}</span>
          </div>
        </div>
      </DialogHeader>
    </DialogContent>
  );
}

function CheckForUpdate() {
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

function Restart({ onClose }: { onClose: () => void }) {
  async function SendRestartRequest() {
    const [code, response] = await GetRequest("restart");

    if (code === 201) {
      toast.info("Restarting", {
        description: "Currently restarting, you might need to refresh the page"
      });

      onClose();
      return;
    }

    toast.warning(response.error);
  }

  return (
    <DialogContent className="sm:max-w-md">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <WarningIcon className="h-5 w-5 text-red-500" />
          Restart Application
        </DialogTitle>
        <DialogDescription>
          This will restart the entire application. Any unsaved changes may be
          lost.
        </DialogDescription>
      </DialogHeader>

      <div className="flex gap-3 justify-end mt-4">
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button variant="destructive" onClick={SendRestartRequest}>
          Restart
        </Button>
      </div>
    </DialogContent>
  );
}

export default function PauseBlockingDialog({
  onClose
}: {
  onClose: () => void;
}) {
  type PausedResponse = {
    paused: boolean;
    timeLeft: number;
  };

  const [pauseTime, setPauseTime] = useState(10);
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
      const [status, response] = await PostRequest("pause", {
        time: pauseTime
      });

      if (status === 200) {
        toast.info(`Paused blocking for ${pauseTime} seconds`);
        const [getStatus, getResponse] = await GetRequest("pause");
        if (getStatus === 200) {
          setPauseStatus(getResponse);
          if (getResponse.paused) {
            setRemainingTime(getResponse.timeLeft);
          }
        }
      } else {
        toast.error("Failed to pause blocking", {
          description: response.error
        });
      }
    } catch (error) {
      toast.error("Error pausing blocking", {
        description: error instanceof Error ? error.message : String(error)
      });
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

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <DialogContent className="sm:max-w-md">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <ClockIcon className="h-5 w-5 text-primary" />
          {pauseStatus?.paused ? "Blocking Paused" : "Pause Blocking"}
        </DialogTitle>
        <DialogDescription>
          {pauseStatus?.paused
            ? "Blocking is currently paused"
            : "Temporarily allow all traffic through"}
        </DialogDescription>
      </DialogHeader>

      {pauseStatus?.paused ? (
        <div className="py-6 space-y-4">
          <div className="flex flex-col items-center space-y-3">
            <div className="text-4xl font-bold tabular-nums">
              {formatTime(remainingTime)}
            </div>
            <p className="text-sm text-muted-foreground">remaining</p>
          </div>

          <Button
            onClick={handleRemovePause}
            disabled={isLoading}
            className="w-full bg-primary/80 hover:bg-primary"
          >
            <PlayCircleIcon size={18} className="mr-2" />
            {isLoading ? "Resuming..." : "Resume Now"}
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Quick Select</label>
            <ToggleGroup
              type="single"
              variant="outline"
              value={String(pauseTime)}
            >
              <ToggleGroupItem value="10" onClick={() => setPauseTime(10)}>
                10s
              </ToggleGroupItem>
              <ToggleGroupItem value="30" onClick={() => setPauseTime(30)}>
                30s
              </ToggleGroupItem>
              <ToggleGroupItem value="60" onClick={() => setPauseTime(60)}>
                1m
              </ToggleGroupItem>
              <ToggleGroupItem value="300" onClick={() => setPauseTime(300)}>
                5m
              </ToggleGroupItem>
              <ToggleGroupItem value="600" onClick={() => setPauseTime(600)}>
                10m
              </ToggleGroupItem>
            </ToggleGroup>
          </div>

          <div className="space-y-2">
            <label htmlFor="pause-time" className="text-sm font-medium">
              Custom (seconds)
            </label>
            <Input
              id="pause-time"
              type="number"
              min={1}
              value={pauseTime}
              onChange={(e) => setPauseTime(e.target.valueAsNumber)}
            />
          </div>

          <div className="flex gap-2 pt-2">
            <Button variant="outline" onClick={onClose} className="flex-1">
              Cancel
            </Button>
            <Button
              onClick={handlePause}
              disabled={isLoading}
              className="flex-1 bg-primary/80 hover:bg-primary"
            >
              {isLoading ? "Pausing..." : "Pause"}
            </Button>
          </div>
        </div>
      )}
    </DialogContent>
  );
}

export function NavActions() {
  const [isOpen, setIsOpen] = useState(false);
  const [DialogComponent, setDialogComponent] = useState<
    null | ((props: { onClose: () => void }) => JSX.Element)
  >(null);

  const closeDialog = () => {
    setDialogComponent(null);
  };

  return (
    <div>
      <Popover open={isOpen} onOpenChange={setIsOpen}>
        <PopoverTrigger asChild className="cursor-pointer">
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
                            className="cursor-pointer"
                            onClick={() => {
                              setIsOpen(false);
                              setDialogComponent(() => item.dialog);
                            }}
                          >
                            <item.icon className={item.color} />{" "}
                            <span>{item.label}</span>
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
          <DialogComponent onClose={closeDialog} />
        </Dialog>
      )}
    </div>
  );
}
