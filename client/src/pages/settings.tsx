"use client";

import { Combobox } from "@/components/combobox";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { GetRequest, PostRequest, PutRequest } from "@/util";
import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Warning } from "@phosphor-icons/react";

const SETTINGS_SECTIONS = [
  {
    title: "Security",
    description: "Configure security settings",
    settings: []
  },
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
        widgetType: Combobox
      }
    ]
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
        widgetType: Combobox
      },
      {
        label: "Statistics Retention",
        key: "statisticsRetention",
        explanation: "Days to retain system statistics",
        options: [1, 7, 30, 90],
        default: 7,
        widgetType: Combobox
      },
      {
        label: "Disable Logging",
        key: "disableLogging",
        explanation: "Completely turn off logging",
        options: [true, false],
        default: false,
        widgetType: Switch
      }
    ]
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
        widgetType: Input
      }
    ]
  }
];

function parseLogLevel(level: number | string) {
  if (typeof level === "number") {
    switch (level) {
      case 0:
        return "Debug";
      case 1:
        return "Info";
      case 2:
        return "Warning";
      case 3:
        return "Error";
    }
  } else if (typeof level === "string") {
    switch (level.toUpperCase()) {
      case "Debug":
        return 0;
      case "Info":
        return 1;
      case "Warning":
        return 2;
      case "Error":
        return 3;
    }
  }
}

export function Settings() {
  const [preferences, setPreferences] = useState<
    Record<string, string | boolean | number>
  >({});
  const [isChanged, setIsChanged] = useState(false);
  const [toastShown, setToastShown] = useState(false);
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [passwordData, setPasswordData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: ""
  });
  const [passwordError, setPasswordError] = useState("");
  const originalPreferencesRef = useRef<string>("");
  const toastIdRef = useRef<string | number | null>(null);
  const navigate = useNavigate();

  const fetchSettings = async () => {
    try {
      const [status, response] = await GetRequest("settings");

      if (status === 200 && response) {
        const updatedPreferences = {
          font: localStorage.getItem("font") || "JetBrains Mono",
          logLevel: parseLogLevel(response.settings?.logLevel) || "Info",
          statisticsRetention: response.settings?.statisticsRetention || 7,
          disableLogging: response.settings?.loggingDisabled || false,
          cacheTTL: 60
        };

        setPreferences(updatedPreferences);
        originalPreferencesRef.current = JSON.stringify(updatedPreferences);
      } else {
        console.error("Failed to fetch settings, using defaults");
        const defaultPreferences = {
          font: "JetBrains Mono",
          logLevel: "Info",
          statisticsRetention: 7,
          disableLogging: false,
          cacheTTL: 60
        };
        setPreferences(defaultPreferences);
        originalPreferencesRef.current = JSON.stringify(defaultPreferences);
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
        [key]: typeof prev[key] === "number" ? Number(value) : value
      };

      const currentChecksum = JSON.stringify(newPreferences);
      const hasChanged = currentChecksum !== originalPreferencesRef.current;

      setIsChanged(hasChanged);
      return newPreferences;
    });
  };

  const handleSaveChanges = async () => {
    try {
      await PostRequest("settings", preferences);

      originalPreferencesRef.current = JSON.stringify(preferences);

      setIsChanged(false);
      setToastShown(false);

      if (toastIdRef.current) {
        toast.dismiss(toastIdRef.current);
        toastIdRef.current = null;
      }

      setTimeout(() => {
        toast.success("Settings updated successfully");
      }, 100);
    } catch (error) {
      console.error("Error saving settings", error);
      toast.error("Failed to save settings");
    }
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setPasswordData((prev) => ({
      ...prev,
      [name]: value
    }));

    setPasswordError("");
  };

  const handlePasswordSubmit = async () => {
    setPasswordError("");

    if (!passwordData.currentPassword) {
      setPasswordError("Current password is required");
      return;
    }

    if (!passwordData.newPassword) {
      setPasswordError("New password is required");
      return;
    }

    if (passwordData.newPassword !== passwordData.confirmPassword) {
      setPasswordError("New passwords do not match");
      return;
    }

    try {
      const [status, response] = await PutRequest("password", {
        currentPassword: passwordData.currentPassword,
        newPassword: passwordData.newPassword
      });

      if (status === 200) {
        toast.success("Password updated successfully!");
        setIsPasswordModalOpen(false);

        setPasswordData({
          currentPassword: "",
          newPassword: "",
          confirmPassword: ""
        });

        navigate("/login");
      } else if (status === 400) {
        setPasswordError(response);
      } else {
        setPasswordError("An unexpected error occurred.");
      }
    } catch {
      setPasswordError("Failed to update password. Please try again.");
    }
  };

  useEffect(() => {
    if (!isChanged) {
      setToastShown(false);
    }
  }, [isChanged]);

  useEffect(() => {
    if (isChanged && !toastShown) {
      const id = toast("Unsaved Changes", {
        description: "You have pending configuration updates",
        action: {
          label: "Save Now",
          onClick: handleSaveChanges
        },
        duration: Infinity,
        id: "unsaved-changes-toast"
      });
      toastIdRef.current = id;
      setToastShown(true);
    } else if (!isChanged && toastShown) {
      if (toastIdRef.current) {
        toast.dismiss(toastIdRef.current);
        toastIdRef.current = null;
      }
      setToastShown(false);
    }
  }, [isChanged, toastShown]);

  return (
    <>
      <div
        className="container mx-auto px-4 py-8 space-y-6
        w-full
        md:w-4/5
        lg:w-3/4
        xl:w-1/2"
      >
        {SETTINGS_SECTIONS.map(({ title, description, settings }) => (
          <Card
            key={title}
            className="shadow-sm rounded-xl p-4 md:p-6 space-y-4"
          >
            <div className="border-b pb-3 mb-4">
              <h2 className="text-xl font-semibold">{title}</h2>
              <p className="text-sm text-gray-500">{description}</p>
            </div>

            <div className="space-y-4">
              {title === "Security" && (
                <div
                  key="changepassword"
                  className="flex flex-col md:flex-row
                  justify-between
                  items-start md:items-center
                  space-y-2 md:space-y-0
                  md:space-x-4"
                >
                  <div className="flex-grow">
                    <h3 className="text-base font-medium">Change password</h3>
                    <p className="text-xs text-gray-500 mt-1">
                      Change password used to authenticate with the dashboard
                    </p>
                  </div>
                  <Button
                    onClick={() => setIsPasswordModalOpen(true)}
                    variant="outline"
                    className="w-full md:w-auto"
                  >
                    Change Password
                  </Button>
                </div>
              )}
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
                      <p className="text-xs text-gray-500 mt-1">
                        {explanation}
                      </p>
                    </div>

                    <div className="flex-shrink-0 w-full md:w-auto">
                      <Widget
                        {...(Widget === Combobox
                          ? {
                              value: preferences[key] || "",
                              onChange: (value: string) =>
                                handleSelect(key, value),
                              options,
                              className: "w-full md:w-40"
                            }
                          : Widget === Switch
                          ? {
                              checked: Boolean(preferences[key]),
                              onCheckedChange: (value: boolean) =>
                                handleSelect(key, value)
                            }
                          : Widget === Input
                          ? {
                              value: preferences[key] || "",
                              onChange: (
                                e: React.ChangeEvent<HTMLInputElement>
                              ) => handleSelect(key, e.target.value),
                              placeholder: "Enter Value",
                              className: "w-full md:w-40"
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

      <Dialog open={isPasswordModalOpen} onOpenChange={setIsPasswordModalOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Change Password</DialogTitle>
            <DialogDescription>
              Enter your current password and a new password below.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {passwordError && (
              <div className="flex bg-stone-900 border-1 border-red-700 text-red-700 px-4 py-2 rounded-md text-sm">
                <Warning className="mt-1 mr-2" />
                {passwordError}
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="currentPassword" className="text-sm font-medium">
                Current Password
              </label>
              <Input
                id="currentPassword"
                name="currentPassword"
                type="password"
                value={passwordData.currentPassword}
                onChange={handlePasswordChange}
                placeholder="Enter your current password"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="newPassword" className="text-sm font-medium">
                New Password
              </label>
              <Input
                id="newPassword"
                name="newPassword"
                type="password"
                value={passwordData.newPassword}
                onChange={handlePasswordChange}
                placeholder="Enter your new password"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-medium">
                Confirm New Password
              </label>
              <Input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                value={passwordData.confirmPassword}
                onChange={handlePasswordChange}
                placeholder="Confirm your new password"
              />
            </div>
          </div>

          <DialogFooter className="flex flex-col sm:flex-row sm:justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => {
                setIsPasswordModalOpen(false);
                setPasswordError("");
                setPasswordData({
                  currentPassword: "",
                  newPassword: "",
                  confirmPassword: ""
                });
              }}
            >
              Cancel
            </Button>
            <Button onClick={handlePasswordSubmit}>Update Password</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
