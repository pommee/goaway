"use client";

import { useState, useEffect } from "react";
import { Combobox } from "@/components/combobox";
import { Card } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { PostRequest, GetRequest, PutRequest } from "@/util";
import { useNavigate } from "react-router-dom";

const SETTINGS_SECTIONS = [
  {
    title: "Admin Panel",
    description: "Customize dashboard appearance",
    settings: [
      {
        label: "Font",
        key: "font",
        explanation: "Choose your preferred dashboard font",
        options: ["JetBrains Mono", "Arial", "Times New Roman", "Courier New"],
        default: "JetBrains Mono",
        widgetType: Combobox,
      },
      {
        label: "CurrentPassword",
        key: "currentPassword",
        explanation: "Current password",
        options: [],
        default: "password",
        widgetType: Input,
      },
      {
        label: "Password",
        key: "password",
        explanation: "New password",
        options: [],
        default: "password",
        widgetType: Input,
      },
    ],
  },
  {
    title: "Logging",
    description: "Configure logging preferences",
    settings: [
      {
        label: "Log Level",
        key: "logLevel",
        explanation: "Set the verbosity of system logs",
        options: ["Debug", "Info", "Warning", "Error"],
        default: "Info",
        widgetType: Combobox,
      },
      {
        label: "Statistics Retention",
        key: "statisticsRetention",
        explanation: "Days to retain system statistics",
        options: [1, 7, 30, 90],
        default: 7,
        widgetType: Combobox,
      },
      {
        label: "Disable Logging",
        key: "disableLogging",
        explanation: "Completely turn off logging",
        options: [true, false],
        default: false,
        widgetType: Switch,
      },
    ],
  },
  {
    title: "DNS Server",
    description: "Tune DNS server performance",
    settings: [
      {
        label: "Cache TTL",
        key: "cacheTTL",
        explanation: "Domain resolution cache duration (seconds)",
        options: [30, 60, 120, 300],
        default: 60,
        widgetType: Input,
      },
    ],
  },
];

export default function Settings() {
  const [preferences, setPreferences] = useState<Record<string, any>>({});
  const [isChanged, setIsChanged] = useState(false);
  const navigate = useNavigate();

  const fetchSettings = async () => {
    try {
      const [status, response] = await GetRequest("settings");

      if (status === 200 && response) {
        const updatedPreferences = {
          font: localStorage.getItem("font") || "JetBrains Mono",
          logLevel: response.dns?.LogLevel || "Info",
          statisticsRetention: response.dns?.StatisticsRetention || 7,
          disableLogging: response.dns?.LoggingDisabled || false,
          cacheTTL: response.dns?.CacheTTL || 60,
        };

        setPreferences(updatedPreferences);
      } else {
        console.error("Failed to fetch settings, using defaults");
        setPreferences({
          font: "JetBrains Mono",
          logLevel: "Info",
          statisticsRetention: 7,
          disableLogging: false,
          cacheTTL: 60,
        });
      }
    } catch (error) {
      console.error("Failed to fetch settings:", error);
      toast.error("Could not load settings");
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleSelect = (key: string, value: string | number | boolean) => {
    setPreferences((prev) => {
      const newPreferences = {
        ...prev,
        [key]: typeof prev[key] === "number" ? Number(value) : value,
      };

      setIsChanged(JSON.stringify(newPreferences) !== JSON.stringify(prev));
      return newPreferences;
    });
  };

  const handleSaveChanges = async () => {
    try {
      await PostRequest("settings", preferences);
      setIsChanged(false);

      setTimeout(() => {
        toast.success("Settings updated successfully");
      }, 100);
    } catch (error) {
      console.error("Error saving settings", error);
      toast.error("Failed to save settings");
    }

    const newPassword = preferences.password;
    if (newPassword !== undefined) {
      const [passwordChangeStatus, _] = await PutRequest("password", {
        currentPassword: preferences.currentPassword,
        newPassword: preferences.password,
      });
      if (passwordChangeStatus === 200) {
        toast.success("Updated password!");
        navigate("/login");
      }
    }
  };

  useEffect(() => {
    setPreferences(preferences);
    setIsChanged(false);
  }, [preferences]);

  useEffect(() => {
    if (isChanged) {
      toast("Unsaved Changes", {
        description: "You have pending configuration updates",
        action: {
          label: "Save Now",
          onClick: handleSaveChanges,
        },
        duration: 5000,
      });
    }
  }, [isChanged]);

  return (
    <div
      className="container mx-auto px-4 py-8 space-y-6 
      w-full 
      md:w-4/5 
      lg:w-3/4 
      xl:w-1/2"
    >
      {SETTINGS_SECTIONS.map(({ title, description, settings }) => (
        <Card key={title} className="shadow-sm rounded-xl p-4 md:p-6 space-y-4">
          <div className="border-b pb-3 mb-4">
            <h2 className="text-xl font-semibold">{title}</h2>
            <p className="text-sm text-gray-500">{description}</p>
          </div>

          <div className="space-y-4">
            {settings.map(
              ({ label, key, explanation, options, widgetType: Widget }) => (
                <div
                  key={key}
                  className="flex flex-col md:flex-row 
                    justify-between 
                    items-start md:items-center 
                    space-y-2 md:space-y-0 
                    md:space-x-4"
                >
                  <div className="flex-grow">
                    <h3 className="text-base font-medium">{label}</h3>
                    <p className="text-xs text-gray-500 mt-1">{explanation}</p>
                  </div>

                  <div className="flex-shrink-0 w-full md:w-auto">
                    <Widget
                      {...(Widget === Combobox
                        ? {
                            value: preferences[key] || "",
                            onChange: (value: string) =>
                              handleSelect(key, value),
                            options,
                            className: "w-full md:w-40",
                          }
                        : Widget === Switch
                        ? {
                            checked: Boolean(preferences[key]),
                            onCheckedChange: (value: boolean) =>
                              handleSelect(key, value),
                          }
                        : Widget === Input
                        ? {
                            value: preferences[key] || "",
                            onChange: (
                              e: React.ChangeEvent<HTMLInputElement>
                            ) => handleSelect(key, e.target.value),
                            placeholder: "Enter Value",
                            className: "w-full md:w-40",
                          }
                        : {})}
                    />
                  </div>
                </div>
              )
            )}
          </div>
        </Card>
      ))}
    </div>
  );
}
