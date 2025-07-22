"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { GetRequest } from "@/util";
import { NetworkSlashIcon, WarningIcon } from "@phosphor-icons/react";
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
import { Tabs, TabsList, TabsTrigger } from "../../components/ui/tabs";
import {
  NameType,
  ValueType
} from "recharts/types/component/DefaultTooltipContent";

type TopBlockedDomains = {
  frequency: number;
  hits: number;
  name: string;
};

const CustomTooltip = ({
  active,
  payload
}: TooltipContentProps<ValueType, NameType>) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="bg-accent p-2 rounded-md border">
        <p className="font-medium mb-1 truncate max-w-xs">{data.name}</p>
        <div className="flex flex-col gap-1 text-sm">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-primary mr-2" />
            <span className="text-muted-foreground">Hits:</span>
            <span className="ml-1 font-medium">
              {data.hits.toLocaleString()}
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

const EmptyState = () => (
  <div className="flex flex-col items-center justify-center h-full w-full py-10">
    <div className="mb-4">
      <WarningIcon size={36} className="text-destructive" />
    </div>
    <p className="text-muted-foreground text-sm text-center">
      Blocked domains will appear here when detected
    </p>
  </div>
);

const isNewData = (a: TopBlockedDomains[], b: TopBlockedDomains[]): boolean => {
  if (a.length !== b.length) return false;

  return a.every((item, index) => {
    const other = b[index];
    return (
      item.name === other.name &&
      item.frequency === other.frequency &&
      item.hits === other.hits
    );
  });
};

export default function FrequencyChartBlockedDomains() {
  const [data, setData] = useState<TopBlockedDomains[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [sortBy, setSortBy] = useState<"frequency" | "hits">("frequency");
  const previousDataRef = useRef<TopBlockedDomains[]>([]);

  useEffect(() => {
    async function fetchTopBlockedDomains() {
      try {
        const [, domains] = await GetRequest("topBlockedDomains");

        const formattedData = domains.map((domain: TopBlockedDomains) => ({
          name: domain.name,
          hits: domain.hits,
          frequency: domain.frequency
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

    fetchTopBlockedDomains();
    const interval = setInterval(fetchTopBlockedDomains, 2500);
    return () => clearInterval(interval);
  }, []);

  const sortedData = [...data]
    .sort((a, b) => b[sortBy] - a[sortBy])
    .slice(0, 10);

  const formatDomainName = (name: string) => {
    if (name.length > 20) {
      return name.substring(0, 17) + "...";
    }
    return name;
  };

  return (
    <Card className="h-full overflow-hidden">
      <CardHeader>
        <div className="flex justify-between items-center">
          <CardTitle className="flex text-xl font-bold">
            <NetworkSlashIcon className="mt-1 mr-2" /> Top Blocked Domains
          </CardTitle>
          <Tabs
            value={sortBy}
            onValueChange={(value) => setSortBy(value as "frequency" | "hits")}
          >
            <TabsList>
              <TabsTrigger
                value="frequency"
                className="border-l-0 !bg-accent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-orange-600 rounded-none p-0 m-2"
              >
                Frequency
              </TabsTrigger>
              <TabsTrigger
                value="hits"
                className="border-l-0 !bg-accent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-orange-600 rounded-none p-0 m-2"
              >
                Hits
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
                dataKey="name"
                type="category"
                tick={{ fontSize: 12, textAnchor: "end" }}
                tickLine={false}
                axisLine={false}
                width="auto"
                tickFormatter={formatDomainName}
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
