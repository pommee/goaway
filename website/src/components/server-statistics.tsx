import { useEffect, useState } from "react";
import { Thermometer, Database, Cpu, LucideMemoryStick } from "lucide-react";
import { GetRequest } from "@/util";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { DialogDescription } from "@radix-ui/react-dialog";
import { compare } from "compare-versions";
import confetti from "canvas-confetti";

export type Metrics = {
  cpuTemp: number;
  cpuUsage: number;
  dbSize: number;
  portDNS: number;
  portWebsite: number;
  totalMem: number;
  usedMem: number;
  usedMemPercentage: number;
  version: string;
};

const getColor = (value: number, max: number) => {
  const hue = Math.max(0, 120 - (value / max) * 120);
  return `hsl(${hue}, 100%, 50%)`;
};

function MetricItem({
  label,
  value,
  icon: Icon,
  color
}: {
  label: string;
  value: string | number;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="flex items-center justify-between p-1 text-sm rounded-md">
      <div className="flex items-center">
        <Icon size={16} className="mr-3" />
        <span>{label}</span>
      </div>
      <span className="font-mono font-medium" style={{ color }}>
        {value}
      </span>
    </div>
  );
}

async function checkForUpdate() {
  try {
    localStorage.setItem("lastUpdateCheck", Date.now().toString());
    const response = await fetch(
      "https://api.github.com/repos/pommee/goaway/tags"
    );
    const data = await response.json();
    const latestVersion = data[0].name.replace("v", "");
    localStorage.setItem("latestVersion", latestVersion);
    return latestVersion;
  } catch (error) {
    console.error("Failed to check for updates:", error);
    return null;
  }
}

function shouldCheckForUpdate() {
  const lastUpdateCheck = localStorage.getItem("lastUpdateCheck");
  if (!lastUpdateCheck) {
    return true;
  }
  const lastCheckTime = parseInt(lastUpdateCheck, 10);
  const fiveMinutesInMs = 5 * 60 * 1000;
  return Date.now() - lastCheckTime > fiveMinutesInMs;
}

function celebrateUpdate() {
  const colors = ["#a786ff", "#fd8bbc", "#eca184", "#f8deb1"];
  confetti({
    particleCount: 50,
    angle: 60,
    spread: 55,
    startVelocity: 60,
    origin: { x: 0, y: 0.8 },
    colors: colors
  });
  confetti({
    particleCount: 50,
    angle: 120,
    spread: 55,
    startVelocity: 60,
    origin: { x: 1, y: 0.8 },
    colors: colors
  });
}

export function ServerStatistics() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [updateNotified, setUpdateNotified] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateLogs, setUpdateLogs] = useState<string[]>([]);
  const [newVersion, setNewVersion] = useState<string>("");

  useEffect(() => {
    if (shouldCheckForUpdate()) {
      checkForUpdate();
    }

    const interval = setInterval(() => {
      if (shouldCheckForUpdate()) {
        checkForUpdate();
      }
    }, 60000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    async function fetchData() {
      try {
        const [, data] = await GetRequest("server");
        setMetrics(data);

        const latestVersion = localStorage.getItem("latestVersion");
        const installedVersion = data.version;
        localStorage.setItem("installedVersion", installedVersion);

        if (latestVersion !== null) {
          setNewVersion(latestVersion);
          if (
            installedVersion &&
            !updateNotified &&
            compare(latestVersion, installedVersion, ">")
          ) {
            toast(`New version available: v${latestVersion}`, {
              action: {
                label: "Update",
                onClick: () => setShowUpdateModal(true)
              }
            });
            setUpdateNotified(true);
          }
        }
      } catch {
        return;
      }
    }

    fetchData();
    const interval = setInterval(fetchData, 1000);
    return () => clearInterval(interval);
  }, [updateNotified]);

  function startUpdate() {
    setUpdateLogs(["Starting update..."]);

    const eventSource = new EventSource("api/runUpdate");

    eventSource.onmessage = (event) => {
      setUpdateLogs((logs) => [...logs, event.data]);

      if (event.data.includes("Update successful")) {
        toast.info("Updated!", { description: `Now running v${newVersion}` });
        localStorage.setItem("installedVersion", newVersion);
        setShowUpdateModal(false);
        eventSource.close();
        celebrateUpdate();
      }
    };

    eventSource.onerror = () => {
      setUpdateLogs((logs) => [...logs, "Closing event stream..."]);
      eventSource.close();
    };
  }

  const formatNumber = (num: number) => num.toFixed(1);

  if (!metrics) {
    return (
      <div className="bg-gray-800 rounded-lg m-2 p-4 text-gray-300">
        Loading...
      </div>
    );
  }

  return (
    <>
      <div className="bg-gray-800 rounded-lg m-2 overflow-hidden shadow-lg">
        <div className="border-b border-gray-700 px-4 py-2 flex items-center justify-between">
          <span className="text-xs font-medium text-gray-400">
            Server Status
          </span>
          <span className="text-xs bg-gray-700 px-2 py-1 rounded-full text-gray-300 font-mono">
            v{metrics.version}
          </span>
        </div>

        <div className="p-3 space-y-1">
          <MetricItem
            label="CPU"
            value={`${formatNumber(metrics.cpuUsage)}%`}
            icon={Cpu}
            color={getColor(metrics.cpuUsage, 100)}
          />

          <MetricItem
            label="CPU temp"
            value={`${formatNumber(metrics.cpuTemp)}°`}
            icon={Thermometer}
            color={getColor(metrics.cpuTemp, 80)}
          />

          <MetricItem
            label="Memory"
            value={`${formatNumber(metrics.usedMemPercentage)}%`}
            icon={LucideMemoryStick}
            color={getColor(metrics.usedMemPercentage, 100)}
          />

          <MetricItem
            label="DB Size"
            value={`${formatNumber(metrics.dbSize)}MB`}
            icon={Database}
            color={getColor(metrics.dbSize, 50)}
          />
        </div>
      </div>

      <Dialog open={showUpdateModal} onOpenChange={setShowUpdateModal}>
        <DialogContent className="w-1/2 max-w-none">
          <DialogHeader>
            <DialogTitle>Do you want to update to v{newVersion}?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            You can find changelog{" "}
            <a href="/changelog" className="font-bold font-und">
              here
            </a>
          </DialogDescription>
          <div className="h-64 overflow-auto bg-gray-900 text-green-400 p-2 font-mono text-sm rounded-md border border-gray-700">
            {updateLogs.length > 0 ? (
              updateLogs.map((log, index) => <div key={index}>{log}</div>)
            ) : (
              <div className="text-gray-400">Waiting for update logs...</div>
            )}
          </div>
          <DialogFooter>
            <Button
              onClick={startUpdate}
              className="bg-blue-500 hover:bg-blue-600"
            >
              Start Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
