import { useEffect, useState } from "react";
import { Card } from "./ui/card";
import { LucideIcon, Shield, ShieldX, Users, Database } from "lucide-react";
import clsx from "clsx";
import { GetRequest } from "@/util";

interface MetricsCardProps {
  title: string;
  valueKey: string;
  Icon: LucideIcon;
  bgColor: string;
  type?: "number" | "percentage";
  metricsData: Metrics | null;
}

export type Metrics = {
  allowed: number;
  blocked: number;
  clients: number;
  domainBlockLen: number;
  percentageBlocked: number;
  total: number;
};

export function MetricsCard({
  title,
  valueKey,
  Icon,
  bgColor,
  type = "number",
  metricsData,
}: MetricsCardProps) {
  const value = metricsData?.[valueKey as keyof Metrics];
  const formattedValue =
    type === "percentage" && value !== undefined
      ? `${value.toFixed(1)}%`
      : value?.toLocaleString();

  return (
    <Card className={clsx("relative p-2 rounded-sm w-full", bgColor)}>
      <div className="relative z-10">
        <p className="text-sm text-gray-300">{title}</p>
        <p className="text-2xl font-bold">{formattedValue}</p>
        <p className="text-sm mt-1 text-gray-300">
          {valueKey === "total"}
          {valueKey === "blocked"}
          {valueKey === "percentageBlocked"}
          {valueKey === "domainBlockLen"}
        </p>
      </div>
      <Icon className="absolute right-4 top-1/2 transform -translate-y-1/2 w-14 h-14" />
    </Card>
  );
}

export function MetricsCards() {
  const [metricsData, setMetricsData] = useState<Metrics | null>(null);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const [_, data] = await GetRequest("metrics");
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
        title="Total queries"
        valueKey="total"
        Icon={Shield}
        bgColor="bg-green-800"
        metricsData={metricsData}
      />
      <MetricsCard
        title="Queries blocked"
        valueKey="blocked"
        Icon={ShieldX}
        bgColor="bg-red-800"
        metricsData={metricsData}
      />
      <MetricsCard
        title="Percentage blocked"
        valueKey="percentageBlocked"
        Icon={Users}
        bgColor="bg-blue-800"
        metricsData={metricsData}
      />
      <MetricsCard
        title="Blocked domains"
        valueKey="domainBlockLen"
        Icon={Database}
        bgColor="bg-purple-800"
        metricsData={metricsData}
      />
    </div>
  );
}
