import { GetRequest } from "@/util";
import {
  DatabaseIcon,
  Icon,
  ShieldIcon,
  TrashIcon,
  UsersIcon
} from "@phosphor-icons/react";
import clsx from "clsx";
import { useEffect, useState } from "react";
import { Card } from "./ui/card";

type Metrics = {
  allowed: number;
  blocked: number;
  clients: number;
  domainBlockLen: number;
  percentageBlocked: number;
  total: number;
};

interface MetricsCardProps {
  title: string;
  valueKey: string;
  Icon: Icon;
  bgColor: string;
  type?: "number" | "percentage";
  metricsData: Metrics | null;
  description?: string;
}

function MetricsCard({
  title,
  valueKey,
  Icon,
  bgColor,
  type = "number",
  metricsData,
  description = ""
}: MetricsCardProps) {
  const value = metricsData?.[valueKey as keyof Metrics];

  const formattedValue =
    type === "percentage" && value !== undefined
      ? `${value.toFixed(1)}%`
      : value?.toLocaleString();

  return (
    <Card
      className={clsx("relative p-2 rounded-lg w-full overflow-hidden")}
      style={{
        background: `linear-gradient(to right, #1a1a1a, ${bgColor})`
      }}
    >
      <div className="relative z-10 flex items-center justify-between">
        <div>
          <p className="text-xs font-medium text-gray-200">{title}</p>
          <p className="text-xl font-bold text-white">{formattedValue}</p>
          {description && (
            <p className="text-xs text-gray-300 mt-0.5">{description}</p>
          )}
        </div>
        <Icon className="w-10 h-10 opacity-60" />
      </div>
    </Card>
  );
}

export default function MetricsCards() {
  const [metricsData, setMetricsData] = useState<Metrics | null>(null);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const [, data] = await GetRequest("dnsMetrics");
        setMetricsData(data);
      } catch (error) {
        console.error("Failed to fetch server statistics:", error);
      }
    }

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 1000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <MetricsCard
        title="Total Queries"
        valueKey="total"
        Icon={ShieldIcon}
        bgColor="#166534"
        metricsData={metricsData}
        description="All DNS queries processed"
      />
      <MetricsCard
        title="Queries Blocked"
        valueKey="blocked"
        Icon={TrashIcon}
        bgColor="#991b1b"
        metricsData={metricsData}
        description="Total queries filtered"
      />
      <MetricsCard
        title="Percent Blocked"
        valueKey="percentageBlocked"
        Icon={UsersIcon}
        bgColor="#1e40af"
        type="percentage"
        metricsData={metricsData}
        description="Percentage of blocked queries"
      />
      <MetricsCard
        title="Blocked Domains"
        valueKey="domainBlockLen"
        Icon={DatabaseIcon}
        bgColor="#6b21a8"
        metricsData={metricsData}
        description="Number of domains in blocklist"
      />
    </div>
  );
}
