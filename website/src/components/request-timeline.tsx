"use client";

import { Bar, BarChart, XAxis, YAxis, Tooltip, Legend } from "recharts";
import { useEffect, useState } from "react";

import { ChartConfig, ChartContainer } from "@/components/ui/chart";

const chartConfig = {
  blocked: {
    label: "Blocked",
    color: "red",
  },
  allowed: {
    label: "Allowed",
    color: "green",
  },
} satisfies ChartConfig;

const getTimeSlotLabel = (date: Date) => {
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const interval = Math.floor(minutes / 15) * 15;
  return `${hours}:${String(interval).padStart(2, "0")}`;
};

const generateTimeSlots = () => {
  const slots = [];
  const now = new Date();
  for (let i = 0; i < 24; i++) {
    for (let j = 0; j < 4; j++) {
      const date = new Date(now);
      date.setHours(now.getHours() - (24 - i), j * 15, 0, 0);
      const timeSlot = getTimeSlotLabel(date);
      slots.push({
        timeSlot,
        blocked: 0,
        allowed: 0,
      });
    }
  }
  return slots;
};

export function RequestTimeline() {
  const [chartData, setChartData] = useState(generateTimeSlots());
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchTimestamps() {
      const res = await fetch("/api/queryTimestamps");
      const data = await res.json();

      const now = new Date();
      const twentyFourHoursAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

      const recentQueries = data.queries.filter(
        (query: { timestamp: string | number | Date }) => {
          const timestamp = new Date(query.timestamp);
          return timestamp >= twentyFourHoursAgo;
        }
      );

      const updatedChartData = generateTimeSlots();

      recentQueries.forEach(
        (query: { timestamp: string | number | Date; blocked: any }) => {
          const timestamp = new Date(query.timestamp);
          const timeSlot = getTimeSlotLabel(timestamp);

          const slotData = updatedChartData.find(
            (entry) => entry.timeSlot === timeSlot
          );

          if (slotData) {
            if (query.blocked) {
              slotData.blocked += 1;
            } else {
              slotData.allowed += 1;
            }
          }
        }
      );
      setChartData(updatedChartData);
      setIsLoading(false);
    }

    fetchTimestamps();
    const interval = setInterval(fetchTimestamps, 1000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return <div></div>;
  }

  return (
    <ChartContainer config={chartConfig} className="w-3/4">
      <BarChart data={chartData}>
        <XAxis dataKey="timeSlot" />
        <YAxis />
        <Tooltip />
        <Legend />
        <Bar dataKey="blocked" fill="var(--color-blocked)" radius={4} />
        <Bar dataKey="allowed" fill="var(--color-allowed)" radius={4} />
      </BarChart>
    </ChartContainer>
  );
}
