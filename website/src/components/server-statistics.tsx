import { compare } from "compare-versions";
import { useEffect, useState } from "react";
import { Globe, Thermometer, Server, Database } from "lucide-react";
import { GetRequest } from "@/util";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { DialogDescription } from "@radix-ui/react-dialog";
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

function MetricItem({
  label,
  value,
  icon: Icon,
  color,
}: {
  label: string;
  value: string | number;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="flex items-center justify-between p-1 text-sm rounded-sm">
      <div className="flex items-center">
        <Icon size={16} className={`text-${color}-400 mr-2`} />
        <span>{label}:</span>
      </div>
      <span className="font-mono">{value}</span>
    </div>
  );
}

async function checkForUpdate() {
  const response = await fetch(
    "https://api.github.com/repos/pommee/goaway/tags"
  );
  const data = await response.json();

  const newestVersion = data[0].name.replace("v", "");
  localStorage.setItem("newestVersion", newestVersion);
}

export function ServerStatistics() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [updateNotified, setUpdateNotified] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateLogs, setUpdateLogs] = useState<string[]>([]);
  const [newVersion, setNewVersion] = useState<string>("");

  useEffect(() => {
    const interval = setInterval(() => {
      checkForUpdate();
    }, 300000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    async function fetchData() {
      try {
        let installedVersion = localStorage.getItem("installedVersion");
        const newestVersion = localStorage.getItem("newestVersion");

        const [, data] = await GetRequest("server");
        setMetrics(data);

        if (installedVersion === null) {
          localStorage.setItem("installedVersion", data.version);
          installedVersion = localStorage.getItem("installedVersion");
        }
        if (newestVersion !== null) {
          setNewVersion(newestVersion);
          if (
            installedVersion &&
            !updateNotified &&
            compare(newestVersion, installedVersion, ">")
          ) {
            toast(`New version available: v${newestVersion}`, {
              action: {
                label: "Update",
                onClick: () => setShowUpdateModal(true),
              },
            });
            setUpdateNotified(true);
          }
        }
      } catch (error) {
        console.error("Failed to fetch server statistics:", error);
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
        eventSource.close();
        confetti({
          particleCount: 100,
          spread: 70,
          origin: { y: 0.6 },
        });
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
      <div className="bg-gray-800 rounded-lg m-2 p-4 text-gray-300">
        <div className="text-center text-sm mb-2">v{metrics.version}</div>
        <div className="bg-[rgba(0,0,0,0.2)] md:bg-[rgba(0,0,0,0.2)] p-1 rounded-sm">
          <div className="space-y-2">
            <MetricItem
              label="CPU"
              value={`${formatNumber(metrics.cpuUsage)}%`}
              icon={Globe}
              color="blue"
            />

            <MetricItem
              label="CPU temp"
              value={`${formatNumber(metrics.cpuTemp)}°`}
              icon={Thermometer}
              color="red"
            />

            <MetricItem
              label="Mem"
              value={`${formatNumber(metrics.usedMemPercentage)}%`}
              icon={Server}
              color="purple"
            />

            <MetricItem
              label="Size"
              value={`${formatNumber(metrics.dbSize)}MB`}
              icon={Database}
              color="green"
            />
          </div>
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
