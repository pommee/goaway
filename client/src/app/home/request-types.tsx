"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { NoContent } from "@/shared";
import { GetRequest } from "@/util";
import { SetStateAction, useEffect, useState } from "react";
import {
  Pie,
  PieChart,
  PolarAngleAxis,
  PolarGrid,
  Radar,
  RadarChart
} from "recharts";

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)"
];

type QueryType = {
  count: number;
  queryType: string;
};

export default function RequestTypeChart() {
  const [chartData, setChartData] = useState<QueryType[]>([]);
  const [chartType, setChartType] = useState("radar");

  useEffect(() => {
    async function fetchQueryTypes() {
      try {
        const [, data] = await GetRequest("queryTypes");
        if (!data || !Array.isArray(data)) {
          return;
        }

        const formattedData = data.map((request: QueryType, index: number) => ({
          count: request.count,
          queryType: request.queryType,
          fill: colors[index % colors.length]
        }));

        setChartData(formattedData);
      } catch (error) {
        console.error("Failed to fetch query types:", error);
      }
    }

    fetchQueryTypes();
    const interval = setInterval(fetchQueryTypes, 1000);
    return () => clearInterval(interval);
  }, []);

  const handleChartTypeChange = (value: SetStateAction<string>) => {
    setChartType(value);
  };

  return (
    <Card className="py-2 min-w-80 gap-2">
      <CardHeader className="mx-2">
        <div className="flex items-center justify-between gap-1 w-full">
          <CardTitle className="text-sm">Request Types</CardTitle>
          <Select value={chartType} onValueChange={handleChartTypeChange}>
            <SelectTrigger className="text-xs">
              <SelectValue placeholder="Chart Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="radar">Radar Chart</SelectItem>
              <SelectItem value="pie">Pie Chart</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      {chartData.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-1 px-2">
          {chartData.map((item: QueryType) => (
            <div
              key={item.queryType}
              className="flex items-center gap-1 px-2 py-1 rounded-full bg-muted/30 border border-border/30 hover:bg-muted/50 transition-colors"
            >
              <div
                className="w-2 h-2 rounded-full shadow-sm"
                style={{ backgroundColor: item.fill }}
              />
              <span className="text-xs text-foreground">{item.queryType}</span>
              <span className="text-xs text-muted-foreground">
                {item.count}
              </span>
            </div>
          ))}
        </div>
      )}
      {chartData.length > 0 ? (
        <CardContent className="flex-1 pb-0 px-0 h-[200px]">
          {chartType === "radar" ? (
            <ChartContainer config={{}}>
              <RadarChart data={chartData}>
                <ChartTooltip
                  cursor={false}
                  content={<ChartTooltipContent />}
                />
                <PolarGrid />
                <PolarAngleAxis dataKey="queryType" />
                <Radar
                  dataKey="count"
                  fill="var(--primary)"
                  fillOpacity={0.6}
                  stroke="var(--muted-foreground)"
                  activeDot={{ r: 4 }}
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
                  nameKey="queryType"
                />
              </PieChart>
            </ChartContainer>
          )}
        </CardContent>
      ) : (
        <CardContent className="flex h-[200px] items-center justify-center">
          <NoContent text={"No query types has been identified"} />
        </CardContent>
      )}
    </Card>
  );
}
