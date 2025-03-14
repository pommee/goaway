import { useEffect, useState } from "react";
import { Globe, Thermometer, Server, Database } from "lucide-react";

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
    <div className="flex items-center justify-between bg-[rgba(0,0,0,0.2)] md:bg-[rgba(0,0,0,0.2)] p-1 text-sm rounded-sm">
      <div className="flex items-center">
        <Icon size={16} className={`text-${color}-400 mr-2`} />
        <span>{label}:</span>
      </div>
      <span className="font-mono">{value}</span>
    </div>
  );
}

export function ServerStatistics() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        const res = await fetch("/api/server");
        const data = await res.json();
        setMetrics(data);
      } catch (error) {
        console.error("Failed to fetch server statistics:", error);
      }
    }

    fetchData();
    const interval = setInterval(fetchData, 1000);
    return () => clearInterval(interval);
  }, []);

  const formatNumber = (num: number) => num.toFixed(1);

  if (!metrics) {
    return (
      <div className="bg-gray-800 rounded-lg m-2 p-4 text-gray-300">
        Loading...
      </div>
    );
  }

  return (
    <div className="bg-gray-800 rounded-lg m-2 p-4 text-gray-300">
      <div className="text-center text-sm mb-2">v{metrics.version}</div>

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
  );
}
