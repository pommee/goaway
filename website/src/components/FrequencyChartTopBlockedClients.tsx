"use client";

import { useState, useEffect } from "react";
import {
  Bar,
  BarChart,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Cell,
  Tooltip,
  LabelList
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { GetRequest } from "@/util";
import { Button } from "@/components/ui/button";
import { Warning } from "@phosphor-icons/react";

export type TopBlockedClients = {
  frequency: number;
  requestCount: number;
  client: string;
};

const CustomTooltip = ({ active, payload }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="bg-white dark:bg-gray-800 p-3 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        <p className="font-medium text-gray-900 dark:text-gray-100 mb-1 truncate max-w-xs">
          {data.client}
        </p>
        <div className="flex flex-col gap-1 text-sm">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-indigo-500 mr-2" />
            <span className="text-gray-600 dark:text-gray-300">Requests:</span>
            <span className="ml-1 font-medium">
              {data.requestCount.toLocaleString()}
            </span>
          </div>
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-purple-500 mr-2" />
            <span className="text-gray-600 dark:text-gray-300">Frequency:</span>
            <span className="ml-1 font-medium">
              {data.frequency.toFixed(2)}%
            </span>
          </div>
        </div>
      </div>
    );
  }
  return null;
};

const EmptyState = () => (
  <div className="flex flex-col items-center justify-center h-full w-full py-10">
    <div className="mb-4">
      <Warning size={36} />
    </div>
    <p className="text-gray-500 dark:text-gray-400 text-center">
      No client requests have been made
    </p>
    <p className="text-gray-400 dark:text-gray-500 text-sm mt-1 text-center">
      Client data will appear here when requests are detected
    </p>
  </div>
);

export default function FrequencyChartTopBlockedClients() {
  const [data, setData] = useState<TopBlockedClients[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [sortBy, setSortBy] = useState<"frequency" | "requestCount">(
    "frequency"
  );

  useEffect(() => {
    async function fetchTopBlockedClients() {
      try {
        const [, clients] = await GetRequest("topClients");
        const formattedData = clients.clients.map(
          (client: TopBlockedClients) => ({
            client: client.client,
            requestCount: client.requestCount,
            frequency: client.frequency
          })
        );

        setData(formattedData);
        setIsLoading(false);
      } catch {
        setIsLoading(false);
      }
    }

    fetchTopBlockedClients();
    const interval = setInterval(fetchTopBlockedClients, 1000);
    return () => clearInterval(interval);
  }, []);

  const sortedData = [...data]
    .sort((a, b) => b[sortBy] - a[sortBy])
    .slice(0, 10);

  const formatClientName = (name: string) => {
    if (name.length > 20) {
      return name.substring(0, 17) + "...";
    }
    return name;
  };

  return (
    <Card className="h-full overflow-hidden">
      <CardHeader>
        <div className="flex justify-between items-center">
          <CardTitle className="text-xl font-bold">Top Clients</CardTitle>
          <div className="flex space-x-2">
            <Button
              className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                sortBy === "frequency"
                  ? "bg-stone-800 text-white"
                  : "bg-stone-500 text-white"
              }`}
              onClick={() => setSortBy("frequency")}
            >
              Frequency
            </Button>
            <Button
              className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                sortBy === "requestCount"
                  ? "bg-stone-800 text-white"
                  : "bg-stone-500 text-white"
              }`}
              onClick={() => setSortBy("requestCount")}
            >
              Requests
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="h-[calc(100%-80px)]">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
          </div>
        ) : sortedData.length > 0 ? (
          <ResponsiveContainer>
            <BarChart
              data={sortedData}
              layout="vertical"
              margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
              barSize={16}
            >
              <XAxis
                type="number"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                domain={[0, "dataMax"]}
              />
              <YAxis
                dataKey="client"
                type="category"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                width={100}
                tickFormatter={formatClientName}
                interval={0}
              />
              <Tooltip
                content={<CustomTooltip />}
                cursor={{ fill: "rgba(0, 0, 0, 0.05)" }}
              />
              <Bar dataKey={sortBy} radius={[0, 4, 4, 0]}>
                {sortedData.map((_, index) => (
                  <Cell key={`cell-${index}`} fill="cornflowerblue" />
                ))}
                <LabelList
                  dataKey={sortBy}
                  formatter={(value: number) =>
                    sortBy === "frequency"
                      ? `${value.toFixed(1)}%`
                      : value.toLocaleString()
                  }
                  style={{ fontSize: "14px", fill: "white" }}
                />
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <EmptyState />
        )}
      </CardContent>
    </Card>
  );
}
