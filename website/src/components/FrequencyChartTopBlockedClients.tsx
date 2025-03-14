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

export type TopBlockedClients = {
  frequency: number;
  requestCount: number;
  client: string;
};

export function FrequencyChartTopBlockedClients() {
  const [data, setData] = useState<TopBlockedClients[]>([]);
  const [chartConfig] = useState<ChartConfig>({
    blocked: { label: "Blocked" },
  });

  useEffect(() => {
    async function fetchTopBlockedClients() {
      try {
        const res = await fetch("/api/topClients");
        const clients = await res.json();

        const formattedData = clients.clients.map(
          (client: TopBlockedClients) => ({
            client: client.client,
            requests: client.requestCount,
            frequency: client.frequency,
          })
        );

        setData(formattedData);
      } catch (error) {
        console.error("Failed to fetch logs:", error);
      }
    }

    fetchTopBlockedClients();
    const interval = setInterval(fetchTopBlockedClients, 1000);
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
                  key={item.client}
                  className="flex items-center"
                  style={{ height: "30px" }}
                >
                  {item.client}
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
                  <XAxis dataKey="requests" type="number" hide />
                  <YAxis dataKey="client" type="category" hide />
                  <ChartTooltip
                    cursor={false}
                    content={<ChartTooltipContent hideLabel />}
                  />
                  <Bar
                    dataKey="requests"
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
