"use client";

import { Pie, PieChart } from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { useEffect, useState } from "react";

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

export type QueryType = {
  count: number;
  queryType: string;
};

export function PieChartRequestType() {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    async function fetchQueryTypes() {
      try {
        const res = await fetch("/api/queryTypes");
        const data = await res.json();

        if (!data.queries || !Array.isArray(data.queries)) {
          console.error("Invalid response format:", data);
          return;
        }

        const formattedData = data.queries.map(
          (request: QueryType, index: number) => ({
            count: request.count,
            requestType: request.queryType,
            fill: colors[index % colors.length],
          })
        );

        setChartData(formattedData);
      } catch (error) {
        console.error("Failed to fetch server statistics:", error);
      }
    }

    fetchQueryTypes();
    const interval = setInterval(fetchQueryTypes, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <Card className="w-1/4 border-2">
      <CardHeader className="items-center pb-0">
        <CardTitle>Request Types</CardTitle>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer
          config={{}}
          className="[&_.recharts-pie-label-text]:fill-foreground"
        >
          <PieChart>
            <ChartTooltip content={<ChartTooltipContent hideLabel />} />
            <Pie data={chartData} dataKey="count" label nameKey="queryType" />
          </PieChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
