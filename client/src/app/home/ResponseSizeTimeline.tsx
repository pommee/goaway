import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip
} from "@/components/ui/chart";
import { NoContent } from "@/shared";
import { GetRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  MagnifyingGlassMinusIcon,
  MagnifyingGlassPlusIcon,
  SpinnerIcon
} from "@phosphor-icons/react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@radix-ui/react-select";
import { useCallback, useEffect, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ReferenceArea,
  XAxis,
  YAxis
} from "recharts";

const chartConfig = {
  total: {
    label: "Total",
    color: "hsl(60, 100%, 50%)"
  },
  avg: {
    label: "Average",
    color: "hsl(217, 91%, 60%)"
  },
  max: {
    label: "Maximum",
    color: "hsl(0, 84%, 60%)"
  },
  min: {
    label: "Minimum",
    color: "hsl(142, 71%, 45%)"
  }
};

type ResponseSizeQuery = {
  start: number;
  total_size_bytes: number;
  avg_response_size_bytes: number;
  min_response_size_bytes: number;
  max_response_size_bytes: number;
};

export default function ResponseSizeTimeline() {
  const [chartData, setChartData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [refAreaLeft, setRefAreaLeft] = useState("");
  const [refAreaRight, setRefAreaRight] = useState("");
  const [zoomedData, setZoomedData] = useState([]);
  const [isZoomed, setIsZoomed] = useState(false);
  const [timelineInterval, setTimelineInterval] = useState("2");

  const fetchData = useCallback(async () => {
    try {
      setIsRefreshing(true);
      const [, responseData] = await GetRequest(
        `responseSizeTimestamps?interval=${timelineInterval}`
      );
      const data = responseData.map((q: ResponseSizeQuery) => ({
        interval: q.start,
        timestamp: new Date(q.start).toISOString(),
        total: q.total_size_bytes,
        avg: q.avg_response_size_bytes,
        min: q.min_response_size_bytes,
        max: q.max_response_size_bytes
      }));

      setChartData(data);
      setZoomedData(data);
      setIsLoading(false);
      setIsRefreshing(false);
    } catch {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, [timelineInterval]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchData();
    }, 0);

    const interval = setInterval(fetchData, 10000);
    return () => {
      clearTimeout(timer);
      clearInterval(interval);
    };
  }, [fetchData]);

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

  const formatBytes = (bytes) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  const filteredData = isZoomed ? zoomedData : getFilteredData();

  if (isLoading) {
    return (
      <Card className="w-full">
        <CardContent className="flex items-center justify-center p-6">
          <div className="flex flex-col items-center space-y-2">
            <p className="flex text-sm text-muted-foreground">
              <SpinnerIcon className="animate-spin mt-1 mr-2" /> Loading
              response size data...
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
            <CardTitle className="lg:flex lg:text-xl">
              Response Size Timeline
              <p className="text-sm text-muted-foreground mt-1 lg:ml-4">
                {timelineInterval}-Minute Intervals,{" "}
                {filteredData.length > 0
                  ? "Updated: " +
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
                className="bg-transparent border text-white hover:bg-stone-800"
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
              variant={"outline"}
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
              <div className="mb-1 text-sm text-muted-foreground">
                {!isZoomed && (
                  <div className="flex items-center ml-2">
                    <MagnifyingGlassPlusIcon weight="bold" className="mr-1" />
                    Drag to zoom: Select an area on the chart to zoom in
                  </div>
                )}
              </div>
              <ChartContainer
                config={chartConfig}
                className="aspect-auto h-[200px] w-full"
              >
                <AreaChart
                  data={filteredData}
                  onMouseDown={handleMouseDown}
                  onMouseMove={handleMouseMove}
                  onMouseUp={handleMouseUp}
                >
                  <defs>
                    <linearGradient id="fillTotal" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--color-total)"
                        stopOpacity={0.8}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--color-total)"
                        stopOpacity={0.1}
                      />
                    </linearGradient>
                    <linearGradient id="fillAvg" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--color-avg)"
                        stopOpacity={0.8}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--color-avg)"
                        stopOpacity={0.1}
                      />
                    </linearGradient>
                    <linearGradient id="fillMax" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--color-max)"
                        stopOpacity={0.8}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--color-max)"
                        stopOpacity={0.1}
                      />
                    </linearGradient>
                    <linearGradient id="fillMin" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--color-min)"
                        stopOpacity={0.8}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--color-min)"
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
                    width={60}
                    tickFormatter={(value) => {
                      return formatBytes(value);
                    }}
                  />
                  <ChartTooltip
                    cursor={{
                      stroke: "#d1d5db",
                      strokeWidth: 1,
                      strokeDasharray: "4 4"
                    }}
                    content={({ active, payload, label }) => {
                      if (!active || !payload || !payload.length) return null;

                      const item = filteredData.find(
                        (d) => d.interval === label
                      );
                      const formattedLabel =
                        item && item.timestamp
                          ? new Date(item.timestamp).toLocaleString("en-US", {
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                              hour12: false
                            })
                          : "N/A";

                      return (
                        <div className="bg-background border rounded-sm shadow-lg px-2 py-1">
                          <p className="text-xs mb-2">{formattedLabel}</p>
                          <div className="space-y-1">
                            {payload.map((entry, index) => (
                              <div
                                key={index}
                                className="flex items-center justify-between gap-2 text-sm"
                              >
                                <div className="flex items-center gap-2">
                                  <div
                                    className="w-2.5 h-2.5 rounded-xs"
                                    style={{ backgroundColor: entry.color }}
                                  />
                                  <span className="text-muted-foreground text-xs mr-4">
                                    {chartConfig[entry.dataKey]?.label ||
                                      entry.dataKey}
                                  </span>
                                </div>
                                <span className="font-medium text-right text-xs">
                                  {formatBytes(entry.value)}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    }}
                  />
                  <Area
                    dataKey="total"
                    type="monotone"
                    fill="url(#fillTotal)"
                    stroke="var(--color-total)"
                    strokeWidth={2}
                    stackId="a"
                  />
                  <Area
                    dataKey="min"
                    type="monotone"
                    fill="url(#fillMin)"
                    stroke="var(--color-min)"
                    strokeWidth={2}
                    stackId="b"
                  />
                  <Area
                    dataKey="avg"
                    type="monotone"
                    fill="url(#fillAvg)"
                    stroke="var(--color-avg)"
                    strokeWidth={2}
                    stackId="c"
                  />
                  <Area
                    dataKey="max"
                    type="monotone"
                    fill="url(#fillMax)"
                    stroke="var(--color-max)"
                    strokeWidth={2}
                    stackId="d"
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
                  <ChartLegend content={<ChartLegendContent />} />
                </AreaChart>
              </ChartContainer>
            </CardContent>
          </>
        ) : (
          <CardContent className="flex h-[220px] items-center justify-center">
            <NoContent text={"No requests recorded"} />
          </CardContent>
        )}
      </Card>
    </div>
  );
}
