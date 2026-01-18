"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { NoContent } from "@/shared";
import { GetRequest } from "@/util";
import { UsersIcon } from "@phosphor-icons/react";
import { useEffect, useState, useRef } from "react";
import {
  Bar,
  BarChart,
  Cell,
  LabelList,
  ResponsiveContainer,
  Tooltip,
  TooltipContentProps,
  XAxis,
  YAxis
} from "recharts";
import {
  NameType,
  ValueType
} from "recharts/types/component/DefaultTooltipContent";

type TopBlockedClients = {
  frequency: number;
  requestCount: number;
  client: string;
  clientName: string;
};

const CustomTooltip = ({
  active,
  payload
}: TooltipContentProps<ValueType, NameType>) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload as TopBlockedClients;
    return (
      <div className="bg-accent p-2 rounded-md border">
        <p className="font-medium mb-1 truncate max-w-xs">{data.client}</p>
        <div className="flex flex-col gap-1 text-sm">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-primary mr-2" />
            <span className="text-muted-foreground">Requests:</span>
            <span className="ml-1 font-medium">
              {data.requestCount.toLocaleString()}
            </span>
          </div>
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-primary mr-2" />
            <span className="text-muted-foreground">Frequency:</span>
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

const isNewData = (a: TopBlockedClients[], b: TopBlockedClients[]): boolean => {
  if (a.length !== b.length) return false;

  return a.every((item, index) => {
    const other = b[index];
    return (
      item.client === other.client &&
      item.frequency === other.frequency &&
      item.requestCount === other.requestCount
    );
  });
};

export default function FrequencyChartTopBlockedClients() {
  const [data, setData] = useState<TopBlockedClients[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [sortBy, setSortBy] = useState<"frequency" | "requestCount">(
    "frequency"
  );
  const previousDataRef = useRef<TopBlockedClients[]>([]);

  useEffect(() => {
    async function fetchTopBlockedClients() {
      try {
        const [, clients] = await GetRequest("topClients");
        const formattedData = clients.map((client: TopBlockedClients) => ({
          client:
            client.clientName === "unknown" ? client.client : client.clientName,
          requestCount: client.requestCount,
          frequency: client.frequency
        }));

        if (!isNewData(formattedData, previousDataRef.current)) {
          setData(formattedData);
          previousDataRef.current = formattedData;
        }

        setIsLoading(false);
      } catch {
        setIsLoading(false);
      }
    }

    fetchTopBlockedClients();
    const interval = setInterval(fetchTopBlockedClients, 2500);
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
    <Card className="h-full overflow-hidden gap-0 py-0 pt-2">
      <CardHeader className="px-4 mb-2">
        <div className="flex justify-between items-center">
          <CardTitle className="flex lg:text-xl font-bold">
            <UsersIcon className="mt-1 mr-2" />
            Top Clients
          </CardTitle>
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
                width="auto"
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
          <NoContent text={"No client data to show"} />
        )}
      </CardContent>
    </Card>
  );
}
