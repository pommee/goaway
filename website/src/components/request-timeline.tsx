import {
  Area,
  AreaChart,
  CartesianGrid,
  XAxis,
  YAxis,
  ResponsiveContainer
} from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent
} from "@/components/ui/chart";
import { GetRequest } from "@/util";
import { useEffect, useState } from "react";
import { Button } from "./ui/button";
import { ArrowsClockwise } from "@phosphor-icons/react";

const chartConfig = {
  blocked: {
    label: "Blocked",
    color: "hsl(0, 84%, 60%)"
  },
  allowed: {
    label: "Allowed",
    color: "hsl(142, 71%, 45%)"
  }
};

type Query = {
  t: number;
  b: boolean;
};

export default function RequestTimeline() {
  const [chartData, setChartData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchData = async () => {
    try {
      setIsRefreshing(true);
      const [, data] = await GetRequest("queryTimestamps");

      const now = new Date();
      const twentyFourHoursAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

      const processedData = data.queries
        .filter((query: Query) => {
          const queryTime = new Date(query.t);
          return queryTime >= twentyFourHoursAgo;
        })
        .reduce((acc: any, query: Query) => {
          const timestamp = new Date(query.t);
          const minutes = Math.floor(timestamp.getMinutes() / 2) * 2;
          const intervalTime = new Date(timestamp);
          intervalTime.setMinutes(minutes, 0, 0);

          const intervalKey = intervalTime.toISOString();

          let entry = acc.find((e) => e.interval === intervalKey);
          if (!entry) {
            entry = {
              interval: intervalKey,
              blocked: 0,
              allowed: 0
            };
            acc.push(entry);
          }

          // eslint-disable-next-line @typescript-eslint/no-unused-expressions
          query.b ? entry.blocked++ : entry.allowed++;

          return acc;
        }, [])
        .sort((a, b) => new Date(a.interval) - new Date(b.interval));

      setChartData(processedData);
      setIsLoading(false);
      setIsRefreshing(false);
    } catch (error) {
      console.error("Error fetching data:", error);
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  const getFilteredData = () => {
    if (!chartData.length) return [];

    const now = new Date();
    const twentyFourHoursAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

    return chartData.filter(
      (item) => new Date(item.interval) >= twentyFourHoursAgo
    );
  };

  const filteredData = getFilteredData();

  if (isLoading) {
    return (
      <Card className="w-full">
        <CardContent className="flex items-center justify-center p-6">
          <div className="flex flex-col items-center space-y-2">
            <div className="h-6 w-6 animate-spin rounded-full border-b-2 border-t-2 border-primary"></div>
            <p className="text-sm text-muted-foreground">
              Loading request data...
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="w-full">
      <Card className="overflow-hidden">
        <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
          <div className="grid flex-1 sm:text-left">
            <CardTitle className="text-xl">Request Timeline</CardTitle>
            <p className="text-sm text-muted-foreground">
              2-Minute Intervals,{" "}
              {filteredData.length > 0
                ? "Last Updated: " +
                  new Date().toLocaleString("sv-SE", {
                    month: "short",
                    day: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit"
                  })
                : "No data available"}
            </p>
          </div>
          <>
            <Button
              className="bg-transparent border-1 text-white hover:bg-stone-800"
              onClick={fetchData}
              disabled={isRefreshing}
            >
              <ArrowsClockwise weight="bold" />
              Refresh
            </Button>
          </>
        </CardHeader>

        {filteredData.length > 0 ? (
          <>
            <CardContent className="px-2 pt-0">
              <ChartContainer
                config={chartConfig}
                className="aspect-auto h-[300px] w-full"
              >
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart
                    data={filteredData}
                    margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient
                        id="fillBlocked"
                        x1="0"
                        y1="0"
                        x2="0"
                        y2="1"
                      >
                        <stop
                          offset="5%"
                          stopColor="var(--color-blocked)"
                          stopOpacity={0.8}
                        />
                        <stop
                          offset="95%"
                          stopColor="var(--color-blocked)"
                          stopOpacity={0.1}
                        />
                      </linearGradient>
                      <linearGradient
                        id="fillAllowed"
                        x1="0"
                        y1="0"
                        x2="0"
                        y2="1"
                      >
                        <stop
                          offset="5%"
                          stopColor="var(--color-allowed)"
                          stopOpacity={0.8}
                        />
                        <stop
                          offset="95%"
                          stopColor="var(--color-allowed)"
                          stopOpacity={0.1}
                        />
                      </linearGradient>
                    </defs>
                    <CartesianGrid
                      vertical={false}
                      strokeDasharray="3 3"
                      opacity={0.2}
                    />
                    <XAxis
                      dataKey="interval"
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      minTickGap={40}
                      tickFormatter={(value) => {
                        const date = new Date(value);
                        return date.toLocaleTimeString("sv-SE", {
                          hour: "numeric",
                          minute: "2-digit"
                        });
                      }}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      width={40}
                      tickFormatter={(value) =>
                        value > 999 ? `${(value / 1000).toFixed(1)}k` : value
                      }
                    />
                    <ChartTooltip
                      cursor={{
                        stroke: "#d1d5db",
                        strokeWidth: 1,
                        strokeDasharray: "4 4"
                      }}
                      content={
                        <ChartTooltipContent
                          labelFormatter={(value) => {
                            return new Date(value).toLocaleString("sv-SE", {
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit"
                            });
                          }}
                        />
                      }
                    />
                    <Area
                      dataKey="allowed"
                      type="monotone"
                      fill="url(#fillAllowed)"
                      stroke="var(--color-allowed)"
                      strokeWidth={2}
                      stackId="a"
                    />
                    <Area
                      dataKey="blocked"
                      type="monotone"
                      fill="url(#fillBlocked)"
                      stroke="var(--color-blocked)"
                      strokeWidth={2}
                      stackId="b"
                    />
                    <ChartLegend content={<ChartLegendContent />} />
                  </AreaChart>
                </ResponsiveContainer>
              </ChartContainer>
            </CardContent>
          </>
        ) : (
          <CardContent className="flex h-[300px] items-center justify-center">
            <div className="text-center">
              <p className="text-lg font-medium">No data available</p>
              <p className="text-sm text-muted-foreground">
                No requests recorded in the selected time range
              </p>
            </div>
          </CardContent>
        )}
      </Card>
    </div>
  );
}
