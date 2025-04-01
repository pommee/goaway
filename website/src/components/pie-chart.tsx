"use client";

import {
  Pie,
  PieChart,
  PolarAngleAxis,
  PolarGrid,
  Radar,
  RadarChart
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent
} from "@/components/ui/chart";
import { useEffect, useState } from "react";
import { GetRequest } from "@/util";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)"
];

export type QueryType = {
  count: number;
  queryType: string;
};

export default function RequestTypeChart() {
  const [chartData, setChartData] = useState([]);
  const [chartType, setChartType] = useState("radar");

  useEffect(() => {
    async function fetchQueryTypes() {
      try {
        const [, data] = await GetRequest("queryTypes");
        if (!data.queries || !Array.isArray(data.queries)) {
          console.error("Invalid response format:", data);
          return;
        }

        const formattedData = data.queries.map(
          (request: QueryType, index: number) => ({
            count: request.count,
            requestType: request.queryType,
            fill: colors[index % colors.length]
          })
        );

        setChartData(formattedData);
      } catch (error) {
        console.error("Failed to fetch query types:", error);
      }
    }

    fetchQueryTypes();
    const interval = setInterval(fetchQueryTypes, 1000);
    return () => clearInterval(interval);
  }, []);

  const handleChartTypeChange = (value) => {
    setChartType(value);
  };

  return (
    <Card className="w-3/8">
      <CardHeader className="pb-0">
        <div className="flex items-center justify-between w-full">
          <CardTitle>Request Types</CardTitle>
          <Select value={chartType} onValueChange={handleChartTypeChange}>
            <SelectTrigger>
              <SelectValue placeholder="Chart Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="radar">Radar Chart</SelectItem>
              <SelectItem value="pie">Pie Chart</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        {chartType === "radar" ? (
          <ChartContainer config={{}}>
            <RadarChart data={chartData}>
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent hideLabel />}
              />
              <PolarGrid />
              <PolarAngleAxis dataKey="requestType" />
              <Radar
                dataKey="count"
                fill="#8884d8"
                fillOpacity={0.6}
                stroke="#8884d8"
                activeDot={{ r: 8 }}
              />
            </RadarChart>
          </ChartContainer>
        ) : (
          <ChartContainer
            config={{}}
            className="[&_.recharts-pie-label-text]:fill-foreground"
          >
            <PieChart>
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <Pie
                data={chartData}
                dataKey="count"
                label
                nameKey="requestType"
              />
            </PieChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
