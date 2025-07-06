"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GetRequest } from "@/util";
import { WarningIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import {
  Bar,
  BarChart,
  Cell,
  LabelList,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";

type TopBlockedClients = {
  frequency: number;
  requestCount: number;
  client: string;
};

const CustomTooltip = ({ active, payload }) => {
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
      <WarningIcon size={36} />
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
          <Tabs
            value={sortBy}
            onValueChange={(value) =>
              setSortBy(value as "frequency" | "requestCount")
            }
          >
            <TabsList>
              <TabsTrigger
                value="frequency"
                className="border-l-0 !bg-accent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-orange-600 rounded-none p-0 m-2"
              >
                Frequency
              </TabsTrigger>
              <TabsTrigger
                value="requestCount"
                className="border-l-0 !bg-accent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-orange-600 rounded-none p-0 m-2"
              >
                Requests
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </CardHeader>
      <CardContent className="h-[calc(100%)]">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
          </div>
        ) : sortedData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={sortedData}
              layout="vertical"
              margin={{ right: 50 }}
              barCategoryGap="10%"
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
                tick={{ fontSize: 12, textAnchor: "end" }}
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
              <Bar dataKey={sortBy} radius={[0, 6, 6, 0]} maxBarSize={24}>
                {sortedData.map((_, index) => (
                  <Cell key={`cell-${index}`} fill="cornflowerblue" />
                ))}
                <LabelList
                  dataKey={sortBy}
                  position="right"
                  offset={8}
                  formatter={(value: number) =>
                    sortBy === "frequency"
                      ? `${value.toFixed(1)}%`
                      : value.toLocaleString()
                  }
                  style={{
                    fontSize: "12px",
                    fill: "#616161"
                  }}
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
