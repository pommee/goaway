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

export type TopBlockedClients = {
  frequency: number;
  requestCount: number;
  client: string;
};

export default function FrequencyChartTopBlockedClients() {
  const [data, setData] = useState<TopBlockedClients[]>([]);
  const [chartConfig] = useState<ChartConfig>({
    blocked: { label: "Blocked" },
  });

  useEffect(() => {
    async function fetchTopBlockedClients() {
      try {
        const [, clients] = await GetRequest("topClients");
        const formattedData = clients.clients.map(
          (client: TopBlockedClients) => ({
            client: client.client,
            requestCount: client.requestCount,
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
    <div className="w-1/2 h-[250px]">
      <Card className="h-full">
        <CardHeader>
          <CardTitle>Top Clients</CardTitle>
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
                dataKey="client"
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
