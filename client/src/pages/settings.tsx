"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { GetRequest, PostRequest } from "@/util";
import { Root } from "@/app/settings/types";
import { SecuritySection } from "@/app/settings/SecuritySection";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  CertificateIcon,
  CircuitryIcon,
  DatabaseIcon,
  KeyIcon,
  LockIcon,
  NotificationIcon,
  ShuffleIcon,
  TextAlignCenterIcon
} from "@phosphor-icons/react";
import { DatabaseSection } from "@/app/settings/DatabaseSection";
import { LoggingSection } from "@/app/settings/LoggingSection";
import { AlertsSection } from "@/app/settings/AlertSection";
import { PasswordModal } from "@/app/settings/PasswordModal";
import { ImportModal } from "@/app/settings/ImportModal";
import { APIKeyDialog } from "@/components/APIKeyDialog";
import { AlertDialog } from "@/components/Alert";

export function Settings() {
  const [preferences, setPreferences] = useState<Root>({
    dns: {
      status: { pausedAt: "", pauseTime: "", paused: false },
      address: "0.0.0.0",
      gateway: "192.168.0.1:53",
      cacheTTL: 3600,
      udpSize: 512,
      tls: { enabled: false, cert: "", key: "" },
      upstream: { preferred: "8.8.8.8:53", fallback: ["1.1.1.1:53"] },
      ports: { udptcp: 53, dot: 853, doh: 443 }
    },
    api: {
      port: 8080,
      authentication: true,
      rateLimit: { enabled: true, maxTries: 5, window: 5 }
    },
    logging: { enabled: true, level: 1 },
    misc: {
      inAppUpdate: false,
      statisticsRetention: 7,
      dashboard: true,
      scheduledBlacklistUpdates: true
    }
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
        setPreferences(response);
        originalPrefs.current = JSON.stringify(response);
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
      const updates: Record<string, () => Root> = {
        apiPort: () => ({ ...prev, api: { ...prev.api, port: Number(value) } }),
        authentication: () => ({
          ...prev,
          api: { ...prev.api, authentication: Boolean(value) }
        }),
        dnsAddress: () => ({
          ...prev,
          dns: { ...prev.dns, address: String(value) }
        }),
        dnsGateway: () => ({
          ...prev,
          dns: { ...prev.dns, gateway: String(value) }
        }),
        dnsPort: () => ({
          ...prev,
          dns: {
            ...prev.dns,
            ports: { ...prev.dns.ports, udptcp: Number(value) }
          }
        }),
        dotPort: () => ({
          ...prev,
          dns: { ...prev.dns, ports: { ...prev.dns.ports, dot: Number(value) } }
        }),
        dohPort: () => ({
          ...prev,
          dns: { ...prev.dns, ports: { ...prev.dns.ports, doh: Number(value) } }
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
          dns: { ...prev.dns, tls: { ...prev.dns.tls, cert: String(value) } }
        }),
        tlsKeyFile: () => ({
          ...prev,
          dns: { ...prev.dns, tls: { ...prev.dns.tls, key: String(value) } }
        }),
        logLevel: () => ({
          ...prev,
          logging: { ...prev.logging, level: Number(value) }
        }),
        loggingEnabled: () => ({
          ...prev,
          logging: { ...prev.logging, enabled: Boolean(value) }
        }),
        statisticsRetention: () => ({
          ...prev,
          misc: { ...prev.misc, statisticsRetention: Number(value) }
        }),
        scheduledBlacklistUpdates: () => ({
          ...prev,
          misc: { ...prev.misc, scheduledBlacklistUpdates: Boolean(value) }
        }),
        inAppUpdate: () => ({
          ...prev,
          misc: { ...prev.misc, inAppUpdate: Boolean(value) }
        })
      };

      return updates[key] ? updates[key]() : prev;
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
      await PostRequest("settings", { ...currentPrefs });
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

  const getSettingValue = (key: string): number | string | boolean => {
    const valueMap: Record<string, number | string | boolean> = {
      apiPort: preferences.api.port,
      authentication: preferences.api.authentication,
      dnsAddress: preferences.dns.address,
      dnsGateway: preferences.dns.gateway,
      dnsPort: preferences.dns.ports.udptcp,
      dotPort: preferences.dns.ports.dot,
      dohPort: preferences.dns.ports.doh,
      cacheTTL: preferences.dns.cacheTTL,
      udpSize: preferences.dns.udpSize,
      tlsCertFile: preferences.dns.tls.cert,
      tlsKeyFile: preferences.dns.tls.key,
      logLevel: preferences.logging.level,
      loggingEnabled: preferences.logging.enabled,
      statisticsRetention: preferences.misc.statisticsRetention,
      scheduledBlacklistUpdates: preferences.misc.scheduledBlacklistUpdates,
      inAppUpdate: preferences.misc.inAppUpdate
    };

    return valueMap[key] ?? "";
  };

  if (loading.main) {
    return <div className="container mx-auto py-8 text-center">Loading...</div>;
  }

  return (
    <div>
      <Tabs defaultValue="security">
        <TabsList className="bg-transparent space-x-2">
          <TabsTrigger
            value="security"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <LockIcon />
            Security
          </TabsTrigger>
          <TabsTrigger
            value="api"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <KeyIcon />
            API
          </TabsTrigger>
          <TabsTrigger
            value="logging"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <TextAlignCenterIcon />
            Logging
          </TabsTrigger>
          <TabsTrigger
            value="alerts"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <NotificationIcon />
            Alerts
          </TabsTrigger>
          <TabsTrigger
            value="dns"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <CircuitryIcon />
            DNS
          </TabsTrigger>
          <TabsTrigger
            value="certificate"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <CertificateIcon />
            Certificate
          </TabsTrigger>
          <TabsTrigger
            value="database"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <DatabaseIcon />
            Database
          </TabsTrigger>
          <TabsTrigger
            value="miscellaneous"
            className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
          >
            <ShuffleIcon />
            Miscellaneous
          </TabsTrigger>
        </TabsList>
        <div className="flex flex-col items-center">
          <TabsContent value="security" className="w-1/2">
            <div className="w-full">
              <SecuritySection
                onPasswordClick={() =>
                  setModals((prev) => ({ ...prev, password: true }))
                }
                onApiKeyClick={() =>
                  setModals((prev) => ({ ...prev, apiKey: true }))
                }
              />
            </div>
          </TabsContent>
          <TabsContent value="api"></TabsContent>
          <TabsContent value="logging" className="w-1/2">
            <div className="w-full">
              <LoggingSection
                loggingEnabled={preferences.logging.enabled}
                logLevel={preferences.logging.level}
                statisticsRetention={preferences.misc.statisticsRetention}
                onLoggingEnabledChange={(enabled) =>
                  handleSettingChange("loggingEnabled", enabled)
                }
                onLogLevelChange={(level) => {
                  const levelMap: Record<string, number> = {
                    debug: 0,
                    info: 1,
                    warning: 2,
                    error: 3
                  };
                  handleSettingChange("logLevel", levelMap[level] ?? 1);
                }}
                onStatisticsRetentionChange={(days) =>
                  handleSettingChange("statisticsRetention", Number(days))
                }
              />
            </div>
          </TabsContent>
          <TabsContent value="alerts">
            <AlertsSection
              onConfigureClick={() =>
                setModals((prev) => ({ ...prev, notifications: true }))
              }
            />
          </TabsContent>
          <TabsContent value="dns"></TabsContent>
          <TabsContent value="certificate"></TabsContent>
          <TabsContent value="database" className="w-1/2">
            <div className="w-full">
              <DatabaseSection
                loading={loading}
                setLoading={setLoading}
                fileInput={fileInput}
                setFile={setFile}
                setModals={setModals}
              />
            </div>
          </TabsContent>
          <TabsContent value="miscellaneous"></TabsContent>
        </div>
      </Tabs>

      <PasswordModal
        open={modals.password}
        onClose={() => setModals((prev) => ({ ...prev, password: false }))}
        onSubmit={async () => {
          if (!passwords.current) return setError("Current password required");
          if (!passwords.new) return setError("New password required");
          if (passwords.new !== passwords.confirm)
            return setError("Passwords don't match");

          try {
            const { PutRequest } = await import("@/util");
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
        }}
        passwords={passwords}
        setPasswords={setPasswords}
        error={error}
        setError={setError}
      />
      <ImportModal
        open={modals.importConfirm}
        onClose={() => setModals((prev) => ({ ...prev, importConfirm: false }))}
        onConfirm={async () => {
          if (!file) return;
          setLoading((prev) => ({ ...prev, import: true }));

          try {
            const { getApiBaseUrl } = await import("@/util");
            const formData = new FormData();
            formData.append("database", file);
            const response = await fetch(
              `${getApiBaseUrl()}/api/importDatabase`,
              {
                method: "POST",
                body: formData
              }
            );

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
        }}
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
