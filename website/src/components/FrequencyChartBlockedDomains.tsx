"use client";

import { Bar, BarChart, XAxis, YAxis } from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { GetRequest } from "@/util";
import { useEffect, useState } from "react";

export type TopBlockedDomains = {
  frequency: number;
  hits: number;
  name: string;
};

export function FrequencyChartBlockedDomains() {
  const [data, setData] = useState<TopBlockedDomains[]>([]);
  const [chartConfig] = useState<ChartConfig>({
    blocked: { label: "Blocked" },
  });

  useEffect(() => {
    async function fetchTopBlockedDomains() {
      try {
        const [_, domains] = await GetRequest("topBlockedDomains");
        const formattedData = domains.domains.map(
          (domain: TopBlockedDomains) => ({
            name: domain.name,
            blocked: domain.hits,
            frequency: domain.frequency,
          })
        );

        setData(formattedData);
      } catch (error) {
        console.error("Failed to fetch logs:", error);
      }
    }

    fetchTopBlockedDomains();
    const interval = setInterval(fetchTopBlockedDomains, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="w-1/2 h-[250px]">
      <Card className="h-full">
        <CardHeader>
          <CardTitle>Top Blocked Domains</CardTitle>
        </CardHeader>
        <CardContent className="h-[calc(100%-50px)] flex items-center">
          <ChartContainer config={chartConfig} className="w-full h-full">
            <BarChart
              accessibilityLayer
              data={data}
              layout="vertical"
              className="w-full h-full"
              margin={{
                left: 20,
              }}
            >
              <XAxis type="number" dataKey="frequency" hide />
              <YAxis
                dataKey="name"
                type="category"
                tickLine={false}
                tickMargin={10}
                axisLine={false}
                width={100}
                tickFormatter={(value) => value}
              />
              <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
              <Bar dataKey="frequency" fill="cornflowerblue" radius={5} />
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  );
}
