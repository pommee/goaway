"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { APIKeyDialog } from "@/components/APIKeyDialog";
import { GetRequest, PostRequest, PutRequest, getApiBaseUrl } from "@/util";
import {
  CertificateIcon,
  CircuitryIcon,
  DatabaseIcon,
  DownloadIcon,
  KeyIcon,
  LockIcon,
  NotificationIcon,
  ShuffleIcon,
  SpinnerIcon,
  TextAlignCenterIcon,
  UploadIcon,
  WarningIcon
} from "@phosphor-icons/react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Combobox } from "@/components/combobox";
import { Button } from "@/components/ui/button";
import { Card, CardTitle } from "@/components/ui/card";
import { Root } from "@radix-ui/react-slot";
import { Switch } from "@/components/ui/switch";
import { AlertDialog } from "@/components/Alert";

const parseLogLevel = (level: number | string) => {
  const levels = ["Debug", "Info", "Warning", "Error"];
  return typeof level === "number" ? levels[level] : levels.indexOf(level);
};

export function Settings() {
  const [preferences, setPreferences] = useState<Root>({
    dns: {
      address: "0.0.0.0",
      port: 53,
      dotPort: 853,
      dohPort: 443,
      cacheTTL: 60,
      preferredUpstream: "8.8.8.8:53",
      upstreamDNS: [],
      udpSize: 512,
      status: {
        paused: false,
        pausedAt: "",
        pauseTime: 0
      },
      tlsCertFile: "",
      tlsKeyFile: ""
    },
    db: {
      dbType: "sqlite",
      host: "",
      port: 0,
      database: "",
      ssl: false,
      timeZone: "",
      user: "",
      pass: ""
    },
    api: {
      port: 8080,
      authentication: true
    },
    scheduledBlacklistUpdates: true,
    statisticsRetention: 7,
    loggingEnabled: true,
    logLevel: 1,
    inAppUpdate: false
  });
  const originalPrefs = useRef("");
  const latestPreferences = useRef(preferences);
  const [isChanged, setIsChanged] = useState(false);
  const [modals, setModals] = useState({
    password: false,
    apiKey: false,
    importConfirm: false,
    notifications: false
  });
  const [file, setFile] = useState<File | null>(null);
  const [passwords, setPasswords] = useState({
    current: "",
    new: "",
    confirm: ""
  });
  const [loading, setLoading] = useState({
    main: true,
    import: false,
    export: false
  });
  const [error, setError] = useState("");
  const fileInput = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const [unsavedToastId, setUnsavedToastId] = useState<string | number | null>(
    null
  );

  const fetchSettings = async () => {
    try {
      const [status, response] = await GetRequest("settings");
      if (status === 200) {
        const settings = {
          ...response,
          logLevel: parseLogLevel(response.logLevel)
        };
        setPreferences(settings);
        originalPrefs.current = JSON.stringify(settings);
      }
    } catch {
      toast.error("Failed to load settings");
    } finally {
      setLoading((prev) => ({ ...prev, main: false }));
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleSettingChange = (
    key: string,
    value: number | string | boolean
  ) => {
    setPreferences((prev) => {
      let newPrefs = { ...prev };

      const keyUpdaters = {
        apiPort: () => ({ ...prev, api: { ...prev.api, port: Number(value) } }),
        authentication: () => ({
          ...prev,
          api: { ...prev.api, authentication: value }
        }),

        dnsAddress: () => ({ ...prev, dns: { ...prev.dns, address: value } }),
        dnsPort: () => ({ ...prev, dns: { ...prev.dns, port: Number(value) } }),
        dotPort: () => ({
          ...prev,
          dns: { ...prev.dns, dotPort: Number(value) }
        }),
        dohPort: () => ({
          ...prev,
          dns: { ...prev.dns, dohPort: Number(value) }
        }),
        cacheTTL: () => ({
          ...prev,
          dns: { ...prev.dns, cacheTTL: Number(value) }
        }),
        udpSize: () => ({
          ...prev,
          dns: { ...prev.dns, udpSize: Number(value) }
        }),
        tlsCertFile: () => ({
          ...prev,
          dns: { ...prev.dns, tlsCertFile: value }
        }),
        tlsKeyFile: () => ({
          ...prev,
          dns: { ...prev.dns, tlsKeyFile: value }
        }),

        dbType: () => ({ ...prev, db: { ...prev.db, dbType: value } }),
        dbUser: () => ({ ...prev, db: { ...prev.db, user: value } }),
        dbPassword: () => ({ ...prev, db: { ...prev.db, pass: value } }),
        dbHost: () => ({ ...prev, db: { ...prev.db, host: value } }),
        dbPort: () => ({ ...prev, db: { ...prev.db, port: value } }),
        dbDatabase: () => ({ ...prev, db: { ...prev.db, database: value } }),
        dbSSL: () => ({ ...prev, db: { ...prev.db, ssl: value } }),
        dbTimeZone: () => ({ ...prev, db: { ...prev.db, timeZone: value } }),

        logging: () => ({ ...prev, loggingEnabled: value }),

        default: () => ({ ...prev, [key]: value })
      };

      newPrefs = keyUpdaters[key] ? keyUpdaters[key]() : keyUpdaters.default();
      return newPrefs;
    });
  };

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      const hasChanged = JSON.stringify(preferences) !== originalPrefs.current;
      if (hasChanged !== isChanged) {
        setIsChanged(hasChanged);
      }
    }, 100);

    return () => clearTimeout(timeoutId);
  }, [preferences, isChanged]);

  useEffect(() => {
    latestPreferences.current = preferences;
  }, [preferences]);

  const saveSettingsCallback = useCallback(async () => {
    try {
      const currentPrefs = latestPreferences.current;
      await PostRequest("settings", {
        ...currentPrefs,
        logLevel: parseLogLevel(currentPrefs.logLevel)
      });
      originalPrefs.current = JSON.stringify(currentPrefs);
      setIsChanged(false);
      if (unsavedToastId) {
        toast.dismiss(unsavedToastId);
        setUnsavedToastId(null);
      }
      toast.success("Settings saved");
    } catch {
      toast.error("Failed to save settings");
    }
  }, [unsavedToastId]);

  useEffect(() => {
    if (isChanged && !unsavedToastId) {
      const toastId = toast.info("Unsaved Changes", {
        description: "You have pending changes",
        action: { label: "Save", onClick: saveSettingsCallback },
        closeButton: true,
        duration: Infinity
      });
      setUnsavedToastId(toastId);
    } else if (!isChanged && unsavedToastId) {
      toast.dismiss(unsavedToastId);
      setUnsavedToastId(null);
    }
  }, [isChanged, unsavedToastId, saveSettingsCallback]);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (preferences.db.dbType != "sqlite" || file?.name.endsWith(".db")) {
      setFile(file);
      setModals((prev) => ({ ...prev, importConfirm: true }));
    } else {
      toast.error("Please select a .db file");
    }
  };

  const importDb = async () => {
    if (!file) return;
    setLoading((prev) => ({ ...prev, import: true }));

    try {
      const formData = new FormData();
      formData.append("database", file);
      const response = await fetch(`${getApiBaseUrl()}/api/${preferences.db.dbType}/import`, {
        method: "POST",
        body: formData
      });

      if (response.ok) {
        const result = await response.json();
        toast.success("Database imported", {
          description: result.backup_created
            ? `Backup: ${result.backup_created}`
            : ""
        });
      } else {
        throw new Error(await response.text());
      }
    } catch (err) {
      toast.error("Import failed", {
        description: err instanceof Error ? err.message : ""
      });
    } finally {
      setLoading((prev) => ({ ...prev, import: false }));
      setModals((prev) => ({ ...prev, importConfirm: false }));
      setFile(null);
      if (fileInput.current) fileInput.current.value = "";
    }
  };

  const formatBytes = (bytes: number) => {
    if (!bytes) return "0 B";
    const units = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
  };

  const exportDb = async () => {
    setLoading((prev) => ({ ...prev, export: true }));
    const toastId = toast.loading("Starting export...", {
      description: "Preparing database for export",
      duration: Infinity
    });

    try {
      const response = await fetch(`${getApiBaseUrl()}/api/${preferences.db.dbType}/export`);
      if (!response.ok)
        throw new Error(`Export failed: ${response.statusText}`);
      if (!response.body) throw new Error("ReadableStream not supported");

      const total = parseInt(response.headers.get("Content-Length") || "0", 10);
      const reader = response.body.getReader();
      const chunks = [];
      let received = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        chunks.push(value);
        received += value.length;

        const progressText =
          total > 0
            ? `${Math.round((received / total) * 100)}% (${formatBytes(
                received
              )} / ${formatBytes(total)})`
            : `Downloaded ${formatBytes(received)}`;

        toast.loading("Downloading database...", {
          id: toastId,
          description: progressText,
          duration: Infinity
        });
      }

      let download = "database";
      switch (preferences.db.dbType) {
        case "sqlite":
          download = "database.db";
          break;
        case "postgres":
          download = "database.dump";
          break;
      }

      console.log(download);
      console.log(preferences);

      const blob = new Blob(chunks, { type: "application/octet-stream" });
      const url = URL.createObjectURL(blob);
      const a = Object.assign(document.createElement("a"), {
        href: url,
        download: download
      });
      document.body.appendChild(a).click();
      a.remove();
      URL.revokeObjectURL(url);

      toast.success("Database exported successfully!", {
        id: toastId,
        description: `Downloaded ${formatBytes(received)}`,
        duration: 4000
      });
    } catch (error) {
      toast.error("Export failed", {
        id: toastId,
        description: error.message || "An error occurred during export",
        duration: 5000
      });
    } finally {
      setLoading((prev) => ({ ...prev, export: false }));
    }
  };

  const updatePassword = async () => {
    if (!passwords.current) return setError("Current password required");
    if (!passwords.new) return setError("New password required");
    if (passwords.new !== passwords.confirm)
      return setError("Passwords don't match");

    try {
      const [status, response] = await PutRequest("password", {
        currentPassword: passwords.current,
        newPassword: passwords.new
      });

      if (status === 200) {
        toast.success("Password updated");
        setModals((prev) => ({ ...prev, password: false }));
        navigate("/login");
      } else {
        setError(response || "Error updating password");
      }
    } catch {
      setError("Failed to update password");
    }
  };

  if (loading.main)
    return <div className="container mx-auto py-8 text-center">Loading...</div>;

  return (
    <div className="container mx-auto space-y-4 xl:w-1/2">
      <p className="text-sm text-muted-foreground">
        Settings marked with an asterisk (*) require a full restart to take
        effect.
      </p>
      {SETTINGS_SECTIONS.map(({ title, description, icon, settings }) => (
        <Card key={title} className="p-4 gap-2">
          <CardTitle className="border-b-1 pb-1">
            <div className="flex">
              <div className="mt-1 p-1 mr-2 rounded-lg bg-primary/10 text-primary">
                {icon}
              </div>
              <h2 className="text-xl font-semibold">{title}</h2>
            </div>
            <p className="mt-1 text-sm font-normal text-muted-foreground">
              {description}
            </p>
          </CardTitle>

          <div className="space-y-4">
            {title === "Security" && (
              <>
                <SettingRow
                  title="Change password"
                  description="Update dashboard login password."
                  action={
                    <Button
                      variant="outline"
                      onClick={() =>
                        setModals((prev) => ({ ...prev, password: true }))
                      }
                    >
                      Change Password
                    </Button>
                  }
                />
                <SettingRow
                  title="API Keys"
                  description="Manage programmatic access keys."
                  action={
                    <Button
                      variant="outline"
                      onClick={() =>
                        setModals((prev) => ({ ...prev, apiKey: true }))
                      }
                    >
                      Manage Keys
                    </Button>
                  }
                />
              </>
            )}

            {title === "Database" && (
              <>
                <SettingRow
                  title="Export database"
                  description="Download current database file"
                  action={
                    <Button
                      variant="outline"
                      onClick={exportDb}
                      disabled={loading.import || loading.export}
                    >
                      {loading.export ? (
                        <>
                          <SpinnerIcon className="animate-spin mr-2" />{" "}
                          Exporting...
                        </>
                      ) : (
                        <>
                          <UploadIcon className="mr-2" /> Export
                        </>
                      )}
                    </Button>
                  }
                />
                <SettingRow
                  title="Import database"
                  description="Replace current database (backup created)"
                  action={
                    <>
                      <input
                        ref={fileInput}
                        type="file"
                        accept={preferences.db.dbType == "sqlite" ? ".db" : undefined}
                        onChange={handleFileUpload}
                        className="hidden"
                      />
                      <Button
                        variant="outline"
                        onClick={() => fileInput.current?.click()}
                        disabled={loading.import || loading.export}
                      >
                        {loading.import ? (
                          <>
                            <SpinnerIcon className="animate-spin mr-2" />{" "}
                            Importing...
                          </>
                        ) : (
                          <>
                            <DownloadIcon className="mr-2" /> Import
                          </>
                        )}
                      </Button>
                    </>
                  }
                />
              </>
            )}

            {title === "Alerts" && (
              <>
                <SettingRow
                  title="Configure"
                  description="Set up how you receive notifications for important events."
                  action={
                    <Button
                      variant="outline"
                      onClick={() =>
                        setModals((prev) => ({ ...prev, notifications: true }))
                      }
                    >
                      Open
                    </Button>
                  }
                />
              </>
            )}

            {settings.map(
              ({ label, key, explanation, options, widgetType: Widget }) => {
                const keyMap = {
                  apiPort: preferences?.api.port,
                  authentication: preferences?.api.authentication,

                  dnsAddress: preferences?.dns.address,
                  dnsPort: preferences?.dns.port,
                  dotPort: preferences?.dns.dotPort,
                  dohPort: preferences?.dns.dohPort,
                  cacheTTL: preferences?.dns.cacheTTL,
                  udpSize: preferences?.dns.udpSize,
                  tlsCertFile: preferences?.dns.tlsCertFile,
                  tlsKeyFile: preferences?.dns.tlsKeyFile,

                  logging: preferences?.loggingEnabled
                };

                const value = keyMap[key] ?? preferences[key];

                return (
                  <SettingRow
                    key={key}
                    title={label}
                    description={explanation}
                    action={
                      <Widget
                        {...(Widget === Combobox
                          ? {
                              value: String(value),
                              onChange: (v: string) =>
                                handleSettingChange(key, v),
                              options,
                              className: "w-40"
                            }
                          : Widget === Switch
                            ? {
                                checked: Boolean(value),
                                onCheckedChange: (v: boolean) =>
                                  handleSettingChange(
                                    key === "logging" ? "loggingEnabled" : key,
                                    v
                                  )
                              }
                            : {
                                value: String(value),
                                onChange: (
                                  e: React.ChangeEvent<HTMLInputElement>
                                ) => handleSettingChange(key, e.target.value),
                                placeholder: label,
                                className: "w-40"
                              })}
                      />
                    }
                  />
                );
              }
            )}
          </div>
        </Card>
      ))}

      <PasswordModal
        open={modals.password}
        onClose={() => setModals((prev) => ({ ...prev, password: false }))}
        onSubmit={updatePassword}
        passwords={passwords}
        setPasswords={setPasswords}
        error={error}
        setError={setError}
      />

      <ImportModal
        open={modals.importConfirm}
        onClose={() => setModals((prev) => ({ ...prev, importConfirm: false }))}
        onConfirm={importDb}
        filename={file?.name}
      />

      <APIKeyDialog
        open={modals.apiKey}
        onOpenChange={(open) =>
          setModals((prev) => ({ ...prev, apiKey: open }))
        }
      />

      <AlertDialog
        open={modals.notifications}
        onOpenChange={(open) =>
          setModals((prev) => ({ ...prev, notifications: open }))
        }
      />
    </div>
  );
}

const SettingRow = ({
  title,
  description,
  action
}: {
  title: string;
  description: string;
  action: React.ReactNode;
}) => (
  <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
    <div>
      <h3 className="font-medium">{title}</h3>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
    <div className="w-full md:w-auto">{action}</div>
  </div>
);

const PasswordModal = ({
  open,
  onClose,
  onSubmit,
  passwords,
  setPasswords,
  error,
  setError
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: () => void;
  passwords: { current: string; new: string; confirm: string };
  setPasswords: (p: { current: string; new: string; confirm: string }) => void;
  error: string;
  setError: (e: string) => void;
}) => (
  <Dialog open={open} onOpenChange={onClose}>
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Change Password</DialogTitle>
        <DialogDescription>
          Update your password. You'll be logged out after changing it.
        </DialogDescription>
      </DialogHeader>

      {error && (
        <div className="flex items-center bg-red-900/20 text-red-500 p-2 rounded text-sm">
          <WarningIcon className="mr-2" />
          {error}
        </div>
      )}

      <div className="space-y-4">
        {["current", "new", "confirm"].map((type) => (
          <div key={type} className="space-y-2">
            <label className="text-sm font-medium">
              {type === "current"
                ? "Current Password"
                : type === "new"
                  ? "New Password"
                  : "Confirm Password"}
            </label>
            <Input
              type="password"
              value={passwords[type as keyof typeof passwords]}
              onChange={(e) => {
                setPasswords({ ...passwords, [type]: e.target.value });
                setError("");
              }}
              placeholder={`Enter ${type} password`}
            />
          </div>
        ))}
      </div>

      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          Cancel
        </Button>
        <Button onClick={onSubmit}>Update</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);

const ImportModal = ({
  open,
  onClose,
  onConfirm,
  filename
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  filename?: string;
}) => {
  const [fileDetails, setFileDetails] = useState<{
    name: string;
    size: number;
    lastModified: number;
  } | null>(null);

  useEffect(() => {
    if (filename) {
      const input = document.querySelector(
        "input[type='file']"
      ) as HTMLInputElement;
      const file = input?.files?.[0];
      if (file) {
        setFileDetails({
          name: file.name,
          size: file.size,
          lastModified: file.lastModified
        });
      }
    }
  }, [filename]);

  const formatBytes = (bytes: number) => {
    if (!bytes) return "0 B";
    const units = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
  };

  const formatDate = (timestamp: number) =>
    new Date(timestamp).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false
    });

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>Confirm Import</DialogTitle>
          <DialogDescription>
            <p>
              Replace the current database with <strong>{filename}</strong>?
            </p>
            A backup of your current database will be created.
            {fileDetails && (
              <div className="mt-4 p-2 rounded text-sm">
                <p>
                  <strong>File Details:</strong>
                </p>
                <ul className="mt-1 list-disc ml-4 space-y-1">
                  <li>
                    <strong>Name:</strong> {fileDetails.name}
                  </li>
                  <li>
                    <strong>Size:</strong> {formatBytes(fileDetails.size)}
                  </li>
                  <li>
                    <strong>Last Modified:</strong>{" "}
                    {formatDate(fileDetails.lastModified)}
                  </li>
                </ul>
              </div>
            )}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            className="hover:font-bold transition-all duration-200 bg-destructive/20"
          >
            Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export interface Root {
  dns: Dns;
  api: Api;
  db: Db;
  scheduledBlacklistUpdates: boolean;
  statisticsRetention: number;
  loggingEnabled: boolean;
  logLevel: number;
  inAppUpdate: boolean;
}

export interface Dns {
  address: string;
  port: number;
  dotPort: number;
  dohPort: number;
  cacheTTL: number;
  preferredUpstream: string;
  upstreamDNS: string[];
  udpSize: number;
  status: Status;
  tlsCertFile: string;
  tlsKeyFile: string;
}

export interface Db {
  dbType: string;
  host: string;
  port: number;
  database: string;
  user: string;
  pass: string;
  ssl: boolean;
  timeZone: string;
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

const SETTINGS_SECTIONS = [
  {
    title: "Security",
    description:
      "Manage user authentication and access control for the dashboard and API.",
    icon: <LockIcon />,
    settings: []
  },
  {
    title: "API",
    description:
      "Configure how the API behaves, including its port and authentication requirements.",
    icon: <KeyIcon />,
    settings: [
      {
        label: "Port *",
        key: "apiPort",
        explanation: "Port the API server listens on.",
        default: 8080,
        widgetType: Input
      },
      {
        label: "Authentication *",
        key: "authentication",
        explanation: "Require login credentials to access the dashboard.",
        options: [true, false],
        default: true,
        widgetType: Switch
      }
    ]
  },
  {
    title: "Logging",
    description:
      "Set logging preferences, including verbosity level and data retention.",
    icon: <TextAlignCenterIcon />,
    settings: [
      {
        label: "Log Level",
        key: "logLevel",
        explanation: "Controls the verbosity of system logs.",
        options: ["Debug", "Info", "Warning", "Error"],
        default: "Info",
        widgetType: Combobox
      },
      {
        label: "Statistics Retention",
        key: "statisticsRetention",
        explanation: "Number of days to retain usage statistics.",
        options: [1, 7, 30, 90],
        default: 7,
        widgetType: Combobox
      },
      {
        label: "Logging",
        key: "logging",
        explanation: "Enable or disable logging across the system.",
        options: [true, false],
        default: true,
        widgetType: Switch
      }
    ]
  },
  {
    title: "Alerts",
    description:
      "Configure how the system notifies you about important events.",
    icon: <NotificationIcon />,
    settings: []
  },
  {
    title: "DNS Server",
    description:
      "Manage core DNS server settings, including ports, caching, and buffer size.",
    icon: <CircuitryIcon />,
    settings: [
      {
        label: "Address *",
        key: "dnsAddress",
        explanation: "The network address to bind the DNS server to.",
        default: "0.0.0.0",
        widgetType: Input
      },
      {
        label: "Port *",
        key: "dnsPort",
        explanation: "Port the DNS server listens on.",
        default: 53,
        widgetType: Input
      },
      {
        label: "DoT Port *",
        key: "dotPort",
        explanation: "Port for DNS-over-TLS traffic.",
        default: 853,
        widgetType: Input
      },
      {
        label: "DoH Port *",
        key: "dohPort",
        explanation: "Port for DNS-over-HTTPS traffic.",
        default: 443,
        widgetType: Input
      },
      {
        label: "Cache TTL *",
        key: "cacheTTL",
        explanation: "How long (in seconds) to cache DNS results.",
        default: 60,
        widgetType: Input
      },
      {
        label: "UDP Size *",
        key: "udpSize",
        explanation: "Maximum UDP packet size in bytes.",
        default: 512,
        widgetType: Input
      }
    ]
  },
  {
    title: "Certificate",
    description:
      "Specify TLS certificates used for DoH (dns-over-https) and DoT (dns-over-tls).",
    icon: <CertificateIcon />,
    settings: [
      {
        label: "TLS Certificate *",
        key: "tlsCertFile",
        explanation: "Path to the TLS certificate file.",
        default: "",
        widgetType: Input
      },
      {
        label: "TLS Key *",
        key: "tlsKeyFile",
        explanation: "Path to the TLS private key file.",
        default: "",
        widgetType: Input
      }
    ]
  },
  {
    title: "Database",
    description:
      "Import, export, and manage the internal database used by the application.",
    icon: <DatabaseIcon />,
    settings: []
  },
  {
    title: "Miscellaneous",
    description:
      "Other configurable options that don't fit into a specific category.",
    icon: <ShuffleIcon />,
    settings: [
      {
        label: "Scheduled Blacklist Updates *",
        key: "scheduledBlacklistUpdates",
        explanation: "Automatically update blacklists on a regular schedule.",
        default: true,
        widgetType: Switch
      },
      {
        label: "In App Updates *",
        key: "inAppUpdate",
        explanation:
          "Enable in-app update checks and automatic version management.",
        default: false,
        widgetType: Switch
      }
    ]
  }
];
