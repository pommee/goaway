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
  CaretLineUpIcon,
  CpuIcon,
  DatabaseIcon,
  DownloadIcon,
  HardDriveIcon,
  ThermometerIcon
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

function calculateProgress(logs: string[]): {
  progress: number;
  currentStep: string;
} {
  const steps = [
    { keywords: ["Starting"], label: "Initializing update..." },
    { keywords: ["Loading"], label: "Loading update script..." },
    { keywords: ["Executing"], label: "Executing update script..." },
    { keywords: ["Downloading", "Fetching"], label: "Downloading update..." },
    { keywords: ["Installing", "Extracting"], label: "Installing update..." },
    { keywords: ["Configuring", "Setting up"], label: "Configuring..." },
    { keywords: ["Restarting", "Restart"], label: "Restarting services..." },
    { keywords: ["Update successful", "Complete"], label: "Update complete!" }
  ];

  let currentStepIndex = 0;
  let currentStep = "No update in progress";

  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    const hasStepKeyword = logs.some((log) =>
      step.keywords.some((keyword) =>
        log.toLowerCase().includes(keyword.toLowerCase())
      )
    );

    if (hasStepKeyword) {
      currentStepIndex = i;
      currentStep = step.label;
    }
  }

  const progress =
    logs.length === 0
      ? 0
      : Math.min(((currentStepIndex + 1) / steps.length) * 100, 100);

  return { progress, currentStep };
}

export function ServerStatistics() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [updateNotified, setUpdateNotified] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateLogs, setUpdateLogs] = useState<string[]>([]);
  const [newVersion, setNewVersion] = useState<string>("");
  const [updateStarted, setUpdateStarted] = useState<boolean>(false);
  const [updateProgress, setUpdateProgress] = useState<number>(0);
  const [currentStep, setCurrentStep] = useState<string>("Preparing update...");

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

  useEffect(() => {
    const { progress, currentStep } = calculateProgress(updateLogs);
    setUpdateProgress(progress);
    setCurrentStep(currentStep);
  }, [updateLogs]);

  function startUpdate() {
    setUpdateStarted(true);
    setUpdateProgress(0);
    setCurrentStep("Starting update...");
    const eventSource = new EventSource("api/runUpdate");

    eventSource.onmessage = (event) => {
      setUpdateLogs((logs) => [...logs, event.data]);

      if (event.data.includes("Update successful")) {
        toast.info("Updated!", { description: `Now running v${newVersion}` });
        localStorage.setItem("installedVersion", newVersion);
        eventSource.close();
        celebrateUpdate();
      }
    };

    eventSource.onerror = (e) => {
      const errorMsg =
        e.data ||
        e.message ||
        `EventSource error: ${e.type}` ||
        "Unknown EventSource error";
      setUpdateLogs((logs) => [...logs, `[error] ${errorMsg}`]);
      setUpdateLogs((logs) => [...logs, "[info] Closing event stream..."]);
      eventSource.close();
    };
  }

  const formatNumber = (num: number | undefined) => {
    if (num === undefined || num === null || isNaN(num)) {
      return "0.0";
    }
    return num.toFixed(1);
  };

  const MetricBar = ({
    label,
    value,
    max,
    unit,
    icon: Icon,
    showIcon = true
  }: {
    label: string;
    value: number | undefined;
    max: number;
    unit: string;
    icon: React.ElementType;
    showIcon?: boolean;
  }) => {
    const safeValue = value ?? 0;
    const percentage = Math.min((safeValue / max) * 100, 100);
    const color = getColor(safeValue, max);

    return (
      <div className="flex items-center gap-2 text-xs">
        {showIcon && <Icon size={12} className="text-gray-400 flex-shrink-0" />}
        <div className="flex-1 min-w-0">
          <div className="flex justify-between items-center mb-1">
            <span className="text-gray-300 truncate">{label}</span>
            <span className="font-mono text-gray-100" style={{ color }}>
              {formatNumber(value)}
              {unit}
            </span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-1">
            <div
              className="h-1 rounded-full transition-all duration-300"
              style={{
                width: `${percentage}%`,
                backgroundColor: color
              }}
            />
          </div>
        </div>
      </div>
    );
  };

  return (
    <>
      <div className="bg-gray-800 rounded-lg mx-2 mt-4 overflow-hidden shadow-lg">
        <div className="bg-gray-800 rounded-lg p-3 space-y-3 border border-gray-700">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-gray-400">
              Server Status
            </span>
            <span className="text-xs bg-gray-700 px-2 py-0.5 rounded text-gray-300 font-mono">
              v{metrics?.version}
            </span>
          </div>

          <div className="space-y-2">
            <MetricBar
              label="CPU"
              value={metrics?.cpuUsage}
              max={100}
              unit="%"
              icon={CpuIcon}
            />
            <MetricBar
              label="Temp"
              value={metrics?.cpuTemp}
              max={80}
              unit="°"
              icon={ThermometerIcon}
            />
            <MetricBar
              label="Memory"
              value={metrics?.usedMemPercentage}
              max={100}
              unit="%"
              icon={HardDriveIcon}
            />
            <MetricBar
              label="DB"
              value={metrics?.dbSize}
              max={200}
              unit="MB"
              icon={DatabaseIcon}
            />
          </div>
        </div>
      </div>

      <Dialog open={showUpdateModal} onOpenChange={setShowUpdateModal}>
        <DialogContent className="sm:max-w-6xl rounded-lg border border-gray-800">
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold flex items-center gap-2">
              Update Available: v{newVersion}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
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
            <div className="space-y-3 p-4 bg-gray-900 rounded-md">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-300">{currentStep}</span>
                <span className="text-gray-400 font-mono">
                  {Math.round(updateProgress)}%
                </span>
              </div>
              <div className="w-full bg-gray-700 rounded-full h-2">
                <div
                  className="h-2 rounded-full transition-all duration-500 ease-out bg-blue-500"
                  style={{ width: `${updateProgress}%` }}
                />
              </div>
            </div>

            <div className="h-96 overflow-auto bg-stone-900 p-4 font-mono text-sm rounded-md">
              {updateLogs.length > 0 ? (
                updateLogs.map((log, index) => (
                  <div key={index} className="leading-relaxed">
                    <p
                      className={
                        log.includes("[info]")
                          ? "text-green-400"
                          : log.includes("[error]")
                          ? "text-red-400"
                          : ""
                      }
                    >
                      {log}
                    </p>
                  </div>
                ))
              ) : (
                <div className="flex items-center justify-center h-full text-gray-400">
                  <div className="text-center">
                    <CaretLineUpIcon className="h-8 w-8 animate-pulse mx-auto mb-2" />
                    <p>Waiting for update to start...</p>
                  </div>
                </div>
              )}
            </div>

            <div className="text-sm text-muted-foreground italic">
              You are recommended to backup your data before proceeding with the
              update.
            </div>
          </div>

          <DialogFooter className="flex items-center justify-between sm:justify-end gap-6 pt-2">
            <Button
              disabled={updateStarted}
              onClick={() => setShowUpdateModal(false)}
              className="bg-stone-800 text-white border-1 border-gray-600 hover:bg-gray-600"
            >
              Remind Me Later
            </Button>
            <Button
              onClick={
                updateStarted === false
                  ? startUpdate
                  : () => setShowUpdateModal(false)
              }
              className="bg-blue-500 hover:bg-blue-600 text-white font-medium transition-colors flex items-center gap-2"
            >
              <DownloadIcon className="h-4 w-4" />
              {updateStarted ? <p>Close</p> : <p>Start Update</p>}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
