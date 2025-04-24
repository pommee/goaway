import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent
} from "@/components/ui/chart";
import { GetRequest } from "@/util";

const chartConfig = {
  blocked: {
    label: "Blocked",
    color: "red"
  },
  allowed: {
    label: "Allowed",
    color: "green"
  }
};

export default function RequestTimeline() {
  const [chartData, setChartData] = React.useState([]);
  const [isLoading, setIsLoading] = React.useState(true);

  React.useEffect(() => {
    const fetchData = async () => {
      try {
        const [, data] = await GetRequest("queryTimestamps");

        const now = new Date();
        const twentyFourHoursAgo = new Date(
          now.getTime() - 24 * 60 * 60 * 1000
        );

        const processedData = data.queries
          .filter((query) => {
            const queryTime = new Date(query.timestamp);
            return queryTime >= twentyFourHoursAgo;
          })
          .reduce((acc, query) => {
            const timestamp = new Date(query.timestamp);
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
            query.blocked ? entry.blocked++ : entry.allowed++;

            return acc;
          }, [])
          .sort((a, b) => new Date(a.interval) - new Date(b.interval));

        setChartData(processedData);
        setIsLoading(false);
      } catch {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="w-full">
      <Card>
        <CardHeader className="flex items-center sm:flex-row">
          <div className="grid flex-1 gap-1 text-center sm:text-left">
            <CardTitle>
              Request Timeline (Last 24 Hours, 2-Minute Intervals)
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-0">
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[200px] w-full"
          >
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="fillBlocked" x1="0" y1="0" x2="0" y2="1">
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
                <linearGradient id="fillAllowed" x1="0" y1="0" x2="0" y2="1">
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
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="interval"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const date = new Date(value);
                  return date.toLocaleTimeString("sv-SE", {
                    hour: "numeric",
                    minute: "2-digit"
                  });
                }}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(value) => {
                      return new Date(value).toLocaleString("sv-SE", {
                        month: "short",
                        day: "numeric",
                        hour: "numeric",
                        minute: "2-digit"
                      });
                    }}
                    indicator="dot"
                  />
                }
              />
              <Area
                dataKey="allowed"
                type="natural"
                fill="url(#fillAllowed)"
                stroke="var(--color-allowed)"
                stackId="a"
              />
              <Area
                dataKey="blocked"
                type="natural"
                fill="url(#fillBlocked)"
                stroke="var(--color-blocked)"
                stackId="b"
              />
              <ChartLegend content={<ChartLegendContent />} />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  );
}
