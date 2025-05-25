import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { GetRequest } from "@/util";
import {
  CaretLineUp,
  Cpu,
  Database,
  Download,
  Memory,
  Thermometer
} from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import confetti from "canvas-confetti";
import { compare } from "compare-versions";
import { useEffect, useState } from "react";
import { toast } from "sonner";

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
  commit: string;
  date: string;
};

const getColor = (value: number, max: number) => {
  const hue = Math.max(0, 120 - (value / max) * 120);
  return `hsl(${hue}, 100%, 50%)`;
};

function MetricItem({
  label,
  value,
  icon: Icon,
  color,
  isLoading = false
}: {
  label: string;
  value: string | number;
  icon: React.ElementType;
  color: string;
  isLoading?: boolean;
}) {
  return (
    <div className="flex items-center justify-between p-1 text-sm rounded-md">
      <div className="flex items-center">
        <Icon size={16} className="mr-3" />
        <span>{label}</span>
      </div>
      {isLoading ? (
        <div className="bg-gray-700 rounded w-10 h-4 animate-pulse"></div>
      ) : (
        <span className="font-mono font-medium" style={{ color }}>
          {value}
        </span>
      )}
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
  const [isLoading, setIsLoading] = useState(true);

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
        setIsLoading(false);

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

  return (
    <>
      <div className="bg-gray-800 rounded-lg m-2 overflow-hidden shadow-lg">
        <div className="border-b border-gray-700 px-4 py-2 flex items-center justify-between">
          <span className="text-xs font-medium text-gray-400">
            Server Status
          </span>
          {isLoading ? (
            <div className="bg-gray-700 rounded-full h-4 w-12 animate-pulse"></div>
          ) : (
            <span className="text-xs bg-gray-700 px-2 py-1 rounded-full text-gray-300 font-mono">
              v{metrics?.version}
            </span>
          )}
        </div>

        <div className="p-3 space-y-1">
          <MetricItem
            label="CPU"
            value={
              isLoading ? "0%" : `${formatNumber(metrics?.cpuUsage || 0)}%`
            }
            icon={Cpu}
            color={getColor(metrics?.cpuUsage || 0, 100)}
            isLoading={isLoading}
          />

          <MetricItem
            label="CPU temp"
            value={isLoading ? "0°" : `${formatNumber(metrics?.cpuTemp || 0)}°`}
            icon={Thermometer}
            color={getColor(metrics?.cpuTemp || 0, 80)}
            isLoading={isLoading}
          />

          <MetricItem
            label="Memory"
            value={
              isLoading
                ? "0%"
                : `${formatNumber(metrics?.usedMemPercentage || 0)}%`
            }
            icon={Memory}
            color={getColor(metrics?.usedMemPercentage || 0, 100)}
            isLoading={isLoading}
          />

          <MetricItem
            label="DB Size"
            value={
              isLoading ? "0MB" : `${formatNumber(metrics?.dbSize || 0)}MB`
            }
            icon={Database}
            color={getColor(metrics?.dbSize || 0, 200)}
            isLoading={isLoading}
          />
        </div>
      </div>

      <Dialog open={showUpdateModal} onOpenChange={setShowUpdateModal}>
        <DialogContent className="sm:max-w-4xl rounded-lg border border-gray-200 shadow-lg dark:border-gray-800">
          <DialogHeader className="space-y-2">
            <DialogTitle className="text-xl font-semibold flex items-center gap-2">
              Update Available: v{newVersion}
            </DialogTitle>
            <DialogDescription className="text-gray-600 dark:text-gray-300">
              A new version is available for installation. View the full{" "}
              <a
                href="/changelog"
                target="_blank"
                className="text-blue-500 hover:text-blue-600 underline transition-colors"
              >
                changelog
              </a>{" "}
              for details.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="h-64 overflow-auto bg-gray-900 text-green-400 p-4 font-mono text-sm rounded-md border border-gray-700 shadow-inner">
              {updateLogs.length > 0 ? (
                updateLogs.map((log, index) => (
                  <div key={index} className="leading-relaxed">
                    {log}
                  </div>
                ))
              ) : (
                <div className="flex items-center justify-center h-full text-gray-400">
                  <div className="text-center">
                    <CaretLineUp className="h-8 w-8 animate-pulse mx-auto mb-2" />
                    <p>Waiting for update to start...</p>
                  </div>
                </div>
              )}
            </div>

            <div className="text-sm text-gray-500 dark:text-gray-400 italic">
              You are recommended to backup your data before proceeding with the
              update.
            </div>
          </div>

          <DialogFooter className="flex items-center justify-between sm:justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => setShowUpdateModal(false)}
              className="border-gray-300 hover:bg-gray-100 dark:border-gray-700 dark:hover:bg-gray-800"
            >
              Remind Me Later
            </Button>
            <Button
              onClick={startUpdate}
              className="bg-blue-500 hover:bg-blue-600 text-white font-medium transition-colors flex items-center gap-2"
            >
              <Download className="h-4 w-4" />
              Start Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
