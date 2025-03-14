"use client";

import { Bar, BarChart, XAxis, YAxis } from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
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
        const res = await fetch("/api/topBlockedDomains");
        const domains = await res.json();

        const formattedData = domains.domains.map(
          (domain: TopBlockedDomains) => ({
            name: domain.name,
            blocked: domain.hits,
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
    <div className="w-full h-fit">
      <Card>
        <CardHeader>
          <CardTitle>Top Blocked Domains</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex w-full items-center">
            <div
              className="w-full flex flex-col justify-between"
              style={{ height: `${data.length * 30}px` }}
            >
              {data.map((item) => (
                <div
                  key={item.name}
                  className="flex items-center"
                  style={{ height: "30px" }}
                >
                  {item.name}
                </div>
              ))}
            </div>

            <div className="w-full">
              <ChartContainer
                className="w-full"
                config={chartConfig}
                style={{ height: `${data.length * 30}px` }}
              >
                <BarChart data={data} layout="vertical">
                  <XAxis dataKey="blocked" type="number" hide />
                  <YAxis dataKey="name" type="category" hide />
                  <ChartTooltip
                    cursor={false}
                    content={<ChartTooltipContent hideLabel />}
                  />
                  <Bar
                    dataKey="blocked"
                    layout="vertical"
                    radius={6}
                    barSize={15}
                    fill="hsl(var(--chart-1))"
                    background={{
                      fill: "var(--chart-2)",
                      radius: 6,
                    }}
                  />
                </BarChart>
              </ChartContainer>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
