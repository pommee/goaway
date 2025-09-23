import {
  Area,
  AreaChart,
  CartesianGrid,
  ReferenceArea,
  XAxis,
  YAxis
} from "recharts";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
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
import { GetRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  ChartLineIcon,
  MagnifyingGlassMinusIcon,
  MagnifyingGlassPlusIcon,
  WarningIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { Button } from "../../components/ui/button";

const chartConfig = {
  blocked: {
    label: "Blocked",
    color: "hsl(0, 84%, 60%)"
  },
  allowed: {
    label: "Allowed",
    color: "hsl(142, 71%, 45%)"
  },
  cached: {
    label: "Cached",
    color: "hsl(62, 86%, 55%)"
  }
};

type Query = {
  start: number;
  blocked: boolean;
  cached: boolean;
  allowed: boolean;
};

export default function RequestTimeline() {
  const [chartData, setChartData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [refAreaLeft, setRefAreaLeft] = useState("");
  const [refAreaRight, setRefAreaRight] = useState("");
  const [zoomedData, setZoomedData] = useState([]);
  const [isZoomed, setIsZoomed] = useState(false);
  const [timelineInterval, setTimelineInterval] = useState("2");

  const fetchData = async () => {
    try {
      setIsRefreshing(true);
      const [, responseData] = await GetRequest(
        `queryTimestamps?interval=${timelineInterval}`
      );
      const data = responseData.map((q: Query) => ({
        interval: q.start,
        timestamp: new Date(q.start).toISOString(),
        blocked: q.blocked,
        cached: q.cached,
        allowed: q.allowed
      }));

      setChartData(data);
      setZoomedData(data);
      setIsLoading(false);
      setIsRefreshing(false);
    } catch {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, [timelineInterval]);

  const getFilteredData = () => {
    if (!chartData.length) return [];

    const now = new Date();
    const twentyFourHoursAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

    return chartData.filter(
      (item) => new Date(item.interval) >= twentyFourHoursAgo
    );
  };

  const handleZoomIn = () => {
    if (refAreaLeft === refAreaRight || refAreaRight === "") {
      setRefAreaLeft("");
      setRefAreaRight("");
      return;
    }

    const indexLeft = chartData.findIndex((d) => d.interval === refAreaLeft);
    const indexRight = chartData.findIndex((d) => d.interval === refAreaRight);

    const startIndex = Math.min(indexLeft, indexRight);
    const endIndex = Math.max(indexLeft, indexRight);

    if (startIndex < 0 || endIndex < 0) {
      setRefAreaLeft("");
      setRefAreaRight("");
      return;
    }

    const filteredData = chartData.slice(startIndex, endIndex + 1);
    setZoomedData(filteredData);
    setIsZoomed(true);
    setRefAreaLeft("");
    setRefAreaRight("");
  };

  const handleZoomOut = () => {
    setZoomedData(getFilteredData());
    setIsZoomed(false);
  };

  const handleMouseDown = (e) => {
    if (!e || !e.activeLabel) return;
    setRefAreaLeft(e.activeLabel);
  };

  const handleMouseMove = (e) => {
    if (!refAreaLeft || !e || !e.activeLabel) return;
    setRefAreaRight(e.activeLabel);
  };

  const handleMouseUp = () => {
    if (refAreaLeft && refAreaRight) {
      handleZoomIn();
    }
  };

  const filteredData = isZoomed ? zoomedData : getFilteredData();

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
      <Card className="overflow-hidden py-2 gap-0">
        <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between sm:space-y-0 px-4">
          <div className="grid sm:text-left">
            <CardTitle className="flex text-xl">
              <ChartLineIcon className="mt-1 mr-2" /> Request Timeline
              <p className="text-sm text-muted-foreground mt-1 ml-4">
                {timelineInterval}-Minute Intervals,{" "}
                {filteredData.length > 0
                  ? "Last Updated: " +
                    new Date().toLocaleString("en-US", {
                      month: "short",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                      hour12: false
                    })
                  : "No data available"}
              </p>
            </CardTitle>
          </div>
          <div className="flex gap-2">
            {isZoomed && (
              <Button
                className="bg-transparent border-1 text-white hover:bg-stone-800"
                onClick={handleZoomOut}
              >
                <MagnifyingGlassMinusIcon weight="bold" className="mr-1" />
                Reset Zoom
              </Button>
            )}
            <div>
              <Select
                value={timelineInterval}
                onValueChange={(value) => setTimelineInterval(value)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="2" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">1 minute</SelectItem>
                  <SelectItem value="2">2 minutes</SelectItem>
                  <SelectItem value="5">5 minutes</SelectItem>
                  <SelectItem value="10">10 minutes</SelectItem>
                  <SelectItem value="20">20 minutes</SelectItem>
                  <SelectItem value="30">30 minutes</SelectItem>
                  <SelectItem value="60">1 hour</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Button
              className="bg-transparent border-1 text-white hover:bg-stone-800"
              onClick={fetchData}
              disabled={isRefreshing}
            >
              <ArrowsClockwiseIcon weight="bold" className="mr-1" />
              Refresh
            </Button>
          </div>
        </CardHeader>

        {filteredData.length > 0 ? (
          <>
            <CardContent className="px-2">
              <div className="mb-2 text-sm text-muted-foreground">
                {!isZoomed && (
                  <div className="flex items-center ml-2">
                    <MagnifyingGlassPlusIcon weight="bold" className="mr-1" />
                    Drag to zoom: Select an area on the chart to zoom in
                  </div>
                )}
              </div>
              <ChartContainer config={chartConfig} className="h-[250px] w-full">
                <AreaChart
                  data={filteredData}
                  onMouseDown={handleMouseDown}
                  onMouseMove={handleMouseMove}
                  onMouseUp={handleMouseUp}
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
                    <linearGradient id="fillCached" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--color-cached)"
                        stopOpacity={0.8}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--color-cached)"
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
                    className="select-none"
                    dataKey="interval"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    minTickGap={40}
                    tickFormatter={(value) => {
                      const date = new Date(value);
                      return date.toLocaleTimeString("en-US", {
                        hour: "numeric",
                        minute: "2-digit",
                        hour12: false
                      });
                    }}
                  />
                  <YAxis
                    className="select-none"
                    tickLine={false}
                    axisLine={false}
                    width={45}
                    tickFormatter={(value) => {
                      if (value >= 1_000_000) {
                        return `${(value / 1_000_000).toFixed(1)}m`;
                      } else if (value >= 1_000) {
                        return `${(value / 1_000).toFixed(1)}k`;
                      } else {
                        return value;
                      }
                    }}
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
                          try {
                            const item = filteredData.find(
                              (d) => d.interval === value
                            );
                            if (item && item.timestamp) {
                              return new Date(item.timestamp).toLocaleString(
                                "en-US",
                                {
                                  month: "short",
                                  day: "numeric",
                                  hour: "2-digit",
                                  minute: "2-digit",
                                  hour12: false
                                }
                              );
                            }
                            return "N/A";
                          } catch {
                            return "N/A";
                          }
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
                  <Area
                    dataKey="cached"
                    type="monotone"
                    fill="url(#fillCached)"
                    stroke="var(--color-cached)"
                    strokeWidth={2}
                    stackId="c"
                  />
                  {refAreaLeft && refAreaRight && (
                    <ReferenceArea
                      x1={refAreaLeft}
                      x2={refAreaRight}
                      strokeOpacity={0.3}
                      fill="#8884d8"
                      fillOpacity={0.3}
                    />
                  )}
                  <ChartLegend
                    content={<ChartLegendContent className="p-0" />}
                  />
                </AreaChart>
              </ChartContainer>
            </CardContent>
          </>
        ) : (
          <CardContent className="flex h-[300px] items-center justify-center">
            <div className="flex flex-col items-center justify-center">
              <div className="mb-4">
                <WarningIcon size={36} className="text-destructive" />
              </div>
              <p className="text-sm text-muted-foreground">
                No requests recorded yet
              </p>
            </div>
          </CardContent>
        )}
      </Card>
    </div>
  );
}
