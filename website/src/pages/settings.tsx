"use client";

import { useState } from "react";
import { Combobox } from "@/components/combobox";
import { Card } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";

const adminPanelSettings = [
  {
    label: "Font",
    key: "font",
    explanation: "Font used on the dashboard.",
    options: ["JetBrains mono", "Arial", "Times New Roman", "Courier New"],
    default: "JetBrains mono",
    widgetType: Combobox,
  },
];

const loggingSettings = [
  {
    label: "Log level",
    key: "logLevel",
    explanation:
      "Different log levels produce different amounts of logs. Default is Info.",
    options: ["Debug", "Info", "Warning", "Error"],
    default: "Info",
    widgetType: Combobox,
  },
  {
    label: "Statistics Retention",
    key: "statisticsRetention",
    explanation:
      "Period of time (in days) to keep statistics, i.e logs, recorded requests, clients and more. Default is 7 days.",
    options: [1, 7, 30, 90],
    default: 7,
    widgetType: Combobox,
  },
  {
    label: "Disable logging",
    key: "disableLogging",
    explanation: "Toggle logs in the container. Default is false.",
    options: [true, false],
    default: false,
    widgetType: Switch,
  },
];

const dnsServerSettings = [
  {
    label: "Cache TTL (in seconds)",
    key: "cacheTTL",
    explanation:
      "Once a domain is resolved, it will be cached. Default is 60 seconds if a TTL is not given when resolving.",
    options: [30, 60, 120, 300], // Change `true/false` to numbers
    default: 60,
    widgetType: Input,
  },
];

export default function Settings() {
  const [preferences, setPreferences] = useState(
    Object.fromEntries(
      [...adminPanelSettings, ...loggingSettings, ...dnsServerSettings].map(
        ({ key, default: defaultValue }) => [key, defaultValue]
      )
    )
  );

  const handleSelect = (key: string, value: string | number | boolean) => {
    setPreferences((prev) => ({
      ...prev,
      [key]:
        typeof prev[key] === "number"
          ? Number(value) // Convert back to a number if it was a number
          : value,
    }));
  };

  return (
    <div className="w-full max-w-7xl mx-auto p-6">
      <div className="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-2 gap-4">
        {[
          { title: "Admin Panel", settings: adminPanelSettings },
          { title: "Logging", settings: loggingSettings },
          { title: "DNS Server", settings: dnsServerSettings },
        ].map(({ title, settings }) => (
          <Card key={title} className="shadow-md rounded-lg p-6">
            <h1 className="text-2xl font-semibold">{title}</h1>
            <p className="text-sm text-gray-500">
              {title === "Admin Panel"
                ? "Settings related to the look and feel of the dashboard."
                : title === "Logging"
                ? "Changes related to logging."
                : "DNS server related settings."}
            </p>
            <div className="border-t my-4 border-gray-300"></div>

            <div className="mt-6 space-y-6">
              {settings.map(
                ({ label, key, explanation, options, widgetType: Widget }) => (
                  <div key={key} className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-lg font-medium">{label}</span>
                      <div className="flex gap-3 items-center">
                        <Widget
                          {...(Widget === Combobox
                            ? {
                                value:
                                  typeof preferences[key] === "boolean"
                                    ? String(preferences[key]) // Convert boolean to string
                                    : preferences[key], // Keep string/number as is
                                onChange: (value: string) =>
                                  handleSelect(key, value),
                                options,
                              }
                            : Widget === Switch
                            ? {
                                checked: Boolean(preferences[key]), // Ensure it's a boolean
                                onCheckedChange: (value: boolean) =>
                                  handleSelect(key, value),
                              }
                            : Widget === Input
                            ? {
                                value:
                                  typeof preferences[key] === "boolean"
                                    ? "" // Avoid boolean being passed, fallback to empty string
                                    : preferences[key] !== undefined
                                    ? String(preferences[key]) // Ensure string value
                                    : "",
                                onChange: (
                                  e: React.ChangeEvent<HTMLInputElement>
                                ) => handleSelect(key, e.target.value),
                                type: "number",
                                placeholder: "Enter TTL",
                                className:
                                  "border rounded-md p-2 focus:outline-none focus:ring-2 focus:ring-blue-500",
                              }
                            : {})}
                        />
                      </div>
                    </div>
                    <p className="text-gray-500 text-sm">{explanation}</p>
                  </div>
                )
              )}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
