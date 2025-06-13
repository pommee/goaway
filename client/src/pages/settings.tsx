"use client";

import { APIKeyDialog } from "@/components/APIKeyDialog";
import { Combobox } from "@/components/combobox";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { getApiBaseUrl, GetRequest, PostRequest, PutRequest } from "@/util";
import { UploadIcon, WarningIcon, DownloadIcon } from "@phosphor-icons/react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

const SETTINGS_SECTIONS = [
  {
    title: "Security",
    description: "Configure security settings",
    settings: []
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
        label: "Logging",
        key: "logging",
        explanation: "Toggle logging",
        options: [true, false],
        default: true,
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
  },
  {
    title: "Database",
    description: "Database settings",
    settings: []
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
    switch (level) {
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

const exportDatabase = async () => {
  try {
    const response = await fetch(`${getApiBaseUrl()}/api/exportDatabase`);

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "database.db";
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);

    toast.info("Exported!", { description: `Database has been exported` });
  } catch (error) {
    console.error("Failed to export database:", error);
    toast.error("Could not export database");
  }
};

const importDatabase = async (file: File) => {
  try {
    const formData = new FormData();
    formData.append("database", file);

    const response = await fetch(`${getApiBaseUrl()}/api/importDatabase`, {
      method: "POST",
      body: formData,
      credentials: "include"
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to import database");
    }

    const result = await response.json();
    toast.success("Database imported successfully!", {
      description: result.backup_created
        ? `Backup created: ${result.backup_created}`
        : "Import completed"
    });
  } catch (error) {
    console.error("Failed to import database:", error);
    toast.error("Could not import database", {
      description:
        error instanceof Error ? error.message : "Unknown error occurred"
    });
  }
};

export function Settings() {
  const [preferences, setPreferences] = useState<Settings>({
    dns: {
      port: 53,
      cacheTTL: 360,
      preferredUpstream: "",
      upstreamDNS: [],
      status: {
        paused: false,
        pausedAt: "",
        pauseTime: 0
      }
    },
    api: {
      port: 0,
      authentication: false
    },
    statisticsRetention: 7,
    loggingEnabled: true,
    logLevel: 0
  });
  const [isChanged, setIsChanged] = useState(false);
  const [toastShown, setToastShown] = useState(false);
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [isApiKeyModalOpen, setIsApiKeyModalOpen] = useState(false);
  const [isImportConfirmOpen, setIsImportConfirmOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [passwordData, setPasswordData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: ""
  });
  const [passwordError, setPasswordError] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isImporting, setIsImporting] = useState(false);
  const originalPreferencesRef = useRef<string>("");
  const toastIdRef = useRef<string | number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const fetchSettings = async () => {
    setIsLoading(true);
    try {
      const [status, response]: [number, Root] = await GetRequest("settings");
      if (status === 200 && response && response.settings) {
        const settings = response.settings;
        settings.logLevel = parseLogLevel(settings.logLevel);
        setPreferences(settings);
        originalPreferencesRef.current = JSON.stringify(settings);
      }
    } catch (error) {
      console.error("Failed to fetch settings:", error);
      toast.error("Could not load settings");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleSelect = (key: string, value: string | number | boolean) => {
    setPreferences((prev) => {
      if (key === "cacheTTL") {
        return {
          ...prev,
          dns: {
            ...prev.dns,
            cacheTTL: typeof value === "number" ? value : Number(value)
          }
        };
      }

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
      const settings = { ...preferences };
      settings.logLevel = parseLogLevel(settings.logLevel);
      await PostRequest("settings", settings);

      originalPreferencesRef.current = JSON.stringify(settings);

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

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!file.name.toLowerCase().endsWith(".db")) {
        toast.error("Please select a valid database file (*.db)");
        return;
      }
      setSelectedFile(file);
      setIsImportConfirmOpen(true);
    }
  };

  const handleImportConfirm = async () => {
    if (!selectedFile) return;

    setIsImporting(true);
    setIsImportConfirmOpen(false);

    await importDatabase(selectedFile);

    setIsImporting(false);
    setSelectedFile(null);

    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const handleImportCancel = () => {
    setIsImportConfirmOpen(false);
    setSelectedFile(null);

    if (fileInputRef.current) {
      fileInputRef.current.value = "";
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

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8 flex justify-center items-center">
        <div className="text-center">Loading settings...</div>
      </div>
    );
  }

  return (
    <>
      <div
        className="container mx-auto space-y-6
        w-full
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
                <div>
                  <div
                    key="changepassword"
                    className="flex flex-col md:flex-row
                  justify-between mb-4
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

                  <div
                    key="apikey"
                    className="flex flex-col md:flex-row
                    justify-between
                    items-start md:items-center
                    space-y-2 md:space-y-0
                    md:space-x-4"
                  >
                    <div className="flex-grow">
                      <h3 className="text-base font-medium">API Keys</h3>
                      <p className="text-xs text-gray-500 mt-1">
                        Manage API keys for programmatic access to the system
                      </p>
                    </div>
                    <Button
                      onClick={() => setIsApiKeyModalOpen(true)}
                      variant="outline"
                      className="w-full md:w-auto"
                    >
                      Manage Keys
                    </Button>
                  </div>
                </div>
              )}
              {title === "Database" && (
                <div className="space-y-4">
                  <div
                    key="exportdatabase"
                    className="flex flex-col md:flex-row
                    justify-between
                    items-start md:items-center
                    space-y-2 md:space-y-0
                    md:space-x-4"
                  >
                    <div className="flex-grow">
                      <h3 className="text-base font-medium">Export database</h3>
                      <p className="text-xs text-gray-500 mt-1">
                        Will download the database file.
                      </p>
                    </div>
                    <Button
                      onClick={() => exportDatabase()}
                      variant="outline"
                      className="w-full md:w-auto"
                      disabled={isImporting}
                    >
                      <UploadIcon className="w-4 h-4 mr-2" />
                      {isImporting ? "Exporting..." : "Export"}
                    </Button>
                  </div>

                  <div
                    key="importdatabase"
                    className="flex flex-col md:flex-row
                    justify-between
                    items-start md:items-center
                    space-y-2 md:space-y-0
                    md:space-x-4"
                  >
                    <div className="flex-grow">
                      <h3 className="text-base font-medium">Import database</h3>
                      <p className="text-xs text-gray-500 mt-1">
                        Replace current database with uploaded file. A backup
                        will be created automatically.
                      </p>
                    </div>
                    <div className="flex flex-col md:flex-row gap-2 w-full md:w-auto">
                      <input
                        ref={fileInputRef}
                        type="file"
                        accept=".db"
                        onChange={handleFileSelect}
                        className="hidden"
                        id="database-import"
                      />
                      <Button
                        onClick={() => fileInputRef.current?.click()}
                        variant="outline"
                        className="w-full md:w-auto"
                        disabled={isImporting}
                      >
                        <DownloadIcon className="w-4 h-4 mr-2" />
                        {isImporting ? "Importing..." : "Import"}
                      </Button>
                    </div>
                  </div>
                </div>
              )}
              {settings.map(
                ({ label, key, explanation, options, widgetType: Widget }) => {
                  let currentValue =
                    key === "cacheTTL"
                      ? preferences.dns?.cacheTTL
                      : preferences[key];

                  if (key === "logging") {
                    currentValue = preferences.loggingEnabled;
                  }

                  return (
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
                                value: currentValue?.toString() || "",
                                onChange: (value: string) =>
                                  handleSelect(key, value),
                                options,
                                className: "w-full md:w-40"
                              }
                            : Widget === Switch
                              ? {
                                  checked: Boolean(currentValue),
                                  onCheckedChange: (value: boolean) =>
                                    handleSelect(
                                      key === "logging"
                                        ? "loggingEnabled"
                                        : key,
                                      value
                                    )
                                }
                              : Widget === Input
                                ? {
                                    value: currentValue?.toString() || "",
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
                  );
                }
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
                <WarningIcon className="mt-1 mr-2" />
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

      <Dialog open={isImportConfirmOpen} onOpenChange={setIsImportConfirmOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Confirm Database Import</DialogTitle>
            <DialogDescription>
              Are you sure you want to import{" "}
              <strong>{selectedFile?.name}</strong>?<br /> This{" "}
              <strong>will replace</strong> your current database.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <div className="flex bg-orange-800/80 border border-orange-600/80 text-orange-200/80 px-4 py-2 text-sm">
              <WarningIcon className="mt-0.5 mr-2 flex-shrink-0" />
              <div>
                <p className="font-bold">Important:</p>
                <p>
                  Your current database will be backed up automatically before
                  the import. This action cannot be undone without restoring
                  from the backup.
                </p>
              </div>
            </div>
          </div>

          <DialogFooter className="flex flex-col sm:flex-row sm:justify-end gap-2">
            <Button variant="outline" onClick={handleImportCancel}>
              Cancel
            </Button>
            <Button
              onClick={handleImportConfirm}
              variant="destructive"
              className="cursor-pointer hover:bg-red-800"
            >
              Import Database
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <APIKeyDialog
        open={isApiKeyModalOpen}
        onOpenChange={setIsApiKeyModalOpen}
      />
    </>
  );
}

export interface Root {
  settings: Settings;
}

export interface Settings {
  font?: string;
  dns: Dns;
  api: Api;
  statisticsRetention: number;
  loggingEnabled: boolean;
  logLevel: number | string;
}

export interface Dns {
  port: number;
  cacheTTL: number;
  preferredUpstream: string;
  upstreamDNS: string[];
  status: Status;
}

export interface Status {
  paused: boolean;
  pausedAt: string;
  pauseTime: number;
}

export interface Api {
  port: number;
  authentication: boolean;
}
