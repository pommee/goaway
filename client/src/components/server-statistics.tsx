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
  ArrowRightIcon,
  CaretLineUpIcon,
  CheckCircleIcon,
  CopyIcon,
  CpuIcon,
  DatabaseIcon,
  DownloadIcon,
  HardDriveIcon,
  InfoIcon,
  ThermometerIcon
} from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import confetti from "canvas-confetti";
import { compare } from "compare-versions";
import React, { useEffect, useState } from "react";
import { toast } from "sonner";
import { Separator } from "./ui/separator";

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
      {showIcon && (
        <Icon size={12} className="shrink-0 text-muted-foreground" />
      )}
      <div className="flex-1 min-w-0">
        <div className="flex justify-between items-center mb-1">
          <span className="truncate">{label}</span>
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

export function ServerStatistics() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [updateNotified, setUpdateNotified] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateLogs, setUpdateLogs] = useState<string[]>([]);
  const [installedVersion, setInstalledVersion] = useState<string>("");
  const [newVersion, setNewVersion] = useState<string>("");
  const [updateStarted, setUpdateStarted] = useState<boolean>(false);
  const [updateProgress, setUpdateProgress] = useState<number>(0);
  const [inAppUpdate, setInAppUpdate] = useState<boolean>(false);
  const [currentStep, setCurrentStep] = useState<string>("Preparing update...");
  const [copied, setCopied] = useState(false);

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
        setMetrics(data as Metrics);

        const latestVersion = localStorage.getItem("latestVersion");
        const installedVersion = (data as Metrics).version;
        localStorage.setItem("installedVersion", installedVersion);
        setInAppUpdate((data as Metrics & { inAppUpdate: boolean }).inAppUpdate);

        if (latestVersion !== null) {
          setInstalledVersion(installedVersion);
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
    // eslint-disable-next-line react-hooks/set-state-in-effect
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
        (e as MessageEvent).data ||
        (e as ErrorEvent).message ||
        `EventSource error: ${e.type}` ||
        "Unknown EventSource error";
      setUpdateLogs((logs) => [...logs, `[error] ${errorMsg}`]);
      setUpdateLogs((logs) => [...logs, "[info] Closing event stream..."]);
      eventSource.close();
    };
  }

  const dockerCommands = `# Stop and remove the current container
docker stop container-name
docker rm container-name

# Pull and run the new version
docker pull pommee/goaway:version`;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(dockerCommands);
      setCopied(true);
      toast.info("Copied commands to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  return (
    <>
      <div className="rounded-lg mx-2 mt-4 overflow-hidden shadow-lg">
        <div className="rounded-lg p-3 space-y-3 border">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium">Server Status</span>
            <span className="text-xs px-2 py-0.5 rounded font-mono bg-accent">
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
            <div className="flex items-center gap-2 text-xs mt-3">
              <DatabaseIcon
                size={12}
                className="shrink-0 text-muted-foreground"
              />
              <div className="flex-1 min-w-0">
                <div className="flex justify-between items-center">
                  <span>Database</span>
                  <span className="px-2 py-0.5 rounded font-mono text-xs bg-accent">
                    {formatNumber(metrics?.dbSize)}MB
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Dialog open={showUpdateModal} onOpenChange={setShowUpdateModal}>
        <DialogContent className="sm:max-w-6xl rounded-lg border">
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold flex items-center gap-2">
              Update Available:{" "}
              <p className="text-muted-foreground">v{installedVersion} </p>
              <ArrowRightIcon className="text-2xl" />
              <p className="text-green-600">v{newVersion}</p>
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

          {inAppUpdate === true ? (
            <>
              <div className="space-y-4 py-2">
                <div className="space-y-3 p-4 rounded-md">
                  <div className="flex items-center justify-between text-sm">
                    <span>{currentStep}</span>
                    <span className="font-mono">
                      {Math.round(updateProgress)}%
                    </span>
                  </div>
                  <div className="w-full bg-foreground rounded-full h-2">
                    <div
                      className="h-2 rounded-full transition-all duration-500 ease-out bg-blue-500"
                      style={{ width: `${updateProgress}%` }}
                    />
                  </div>
                </div>

                <div className="h-96 overflow-auto bg-secondary p-4 font-mono text-sm rounded-md">
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
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                      <div className="text-center">
                        <CaretLineUpIcon className="h-8 w-8 animate-pulse mx-auto mb-2" />
                        <p>Waiting for update to start...</p>
                      </div>
                    </div>
                  )}
                </div>

                <div className="text-sm text-muted-foreground italic">
                  You are recommended to backup your data before proceeding with
                  the update.
                </div>
              </div>

              <DialogFooter className="flex items-center justify-between sm:justify-end gap-6 pt-2">
                <Button
                  disabled={updateStarted}
                  onClick={() => setShowUpdateModal(false)}
                  className="bg-muted-foreground"
                >
                  Remind Me Later
                </Button>
                <Button
                  onClick={
                    updateStarted === false
                      ? startUpdate
                      : () => setShowUpdateModal(false)
                  }
                  className="bg-blue-500 hover:bg-blue-600 font-medium transition-colors flex items-center gap-2"
                >
                  <DownloadIcon className="h-4 w-4" />
                  {updateStarted ? <p>Close</p> : <p>Start Update</p>}
                </Button>
              </DialogFooter>
            </>
          ) : (
            <div className="flex items-start gap-3">
              <div className="w-full">
                <h3 className="flex font-medium text-orange-400 mb-2 underline">
                  <InfoIcon className="mt-1 mr-2" />
                  In-app updates are disabled
                </h3>
                <div>
                  <p className="text-sm mb-3">
                    If you are running inside a Docker container, use the
                    following commands to update:
                  </p>
                  <div className="relative">
                    <div className="bg-popover rounded-md p-4 font-mono text-sm">
                      <div className="text-muted-foreground space-y-1">
                        <div># Stop and remove the current container</div>
                        <div className="text-green-400">
                          docker stop container-name
                        </div>
                        <div className="text-green-400">
                          docker rm container-name
                        </div>
                        <br />
                        <div className="mt-2">
                          # Pull and run the new version
                        </div>
                        <div className="text-green-400">
                          docker pull pommee/goaway:version
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={handleCopy}
                      className="absolute top-2 right-2 p-2 bg-input rounded-md transition-colors duration-500 hover:text-green-500 cursor-pointer"
                      title="Copy Docker commands"
                    >
                      {copied ? (
                        <CheckCircleIcon className="w-4 h-4 text-green-400" />
                      ) : (
                        <CopyIcon className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                  <p className="text-xs text-muted-foreground mt-4">
                    Replace{" "}
                    <code className="bg-accent px-1 py-0.5 rounded">
                      container-name
                    </code>{" "}
                    and{" "}
                    <code className="bg-accent px-1 py-0.5 rounded">
                      version
                    </code>{" "}
                    with your actual values.
                  </p>
                </div>

                <Separator className="my-5" />

                <div>
                  <p className="text-sm mb-3">
                    If you are running the standalone binary, use the following
                    commands to update:
                  </p>
                  <div className="bg-popover rounded-md p-4 font-mono text-sm overflow-x-auto">
                    <div className="space-y-1">
                      <div className="text-muted-foreground">
                        # Stop the binary from running and run one of the
                        following:
                      </div>
                      <br />
                      <div className="text-muted-foreground">
                        # Will install the latest version of goaway into
                        ~/$USER/.local/bin
                      </div>
                      <div className="text-green-400">
                        curl
                        https://raw.githubusercontent.com/pommee/goaway/main/installer.sh
                        | sh
                      </div>
                      <div className="text-muted-foreground">
                        # It's also possible to specify what version to install
                      </div>
                      <div className="text-green-400">
                        curl
                        https://raw.githubusercontent.com/pommee/goaway/main/installer.sh
                        | sh /dev/stdin 0.40.4
                      </div>
                      <br />
                      <div className="text-muted-foreground">
                        # Will replace your goaway binary in the current working
                        directory with the latest version
                      </div>
                      <div className="text-green-400">
                        https://raw.githubusercontent.com/pommee/goaway/refs/heads/main/updater.sh
                        | bash
                      </div>
                    </div>
                  </div>
                </div>
                <p className="text-xs italic text-muted-foreground mt-2">
                  This information might be outdated, in that case refer to the{" "}
                  <a
                    className="font-bold underline"
                    href="https://github.com/pommee/goaway/blob/main/README.md"
                    target="_blank"
                  >
                    repository README
                  </a>
                </p>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
