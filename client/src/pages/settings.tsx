"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { GetRequest, PostRequest } from "@/util";
import { Card, CardTitle } from "@/components/ui/card";
import { APIKeyDialog } from "@/components/APIKeyDialog";
import { AlertDialog } from "@/components/Alert";
import { Root } from "@/app/settings/types";
import { SETTINGS_SECTIONS } from "@/app/settings/SettingsSection";
import { DatabaseSection } from "@/app/settings/DatabaseSection";
import { AlertsSection } from "@/app/settings/AlertSection";
import { DynamicSettingsSection } from "@/app/settings/DynamicSettingsSection";
import { SecuritySection } from "@/app/settings/SecuritySection";
import { PasswordModal } from "@/app/settings/PasswordModal";
import { ImportModal } from "@/app/settings/ImportModal";
import { LoggingSection } from "@/app/settings/LoggingSection";

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
    <div className="container mx-auto space-y-4 xl:w-1/2">
      <p className="text-sm text-muted-foreground">
        Settings marked with an asterisk (*) require a full restart to take
        effect.
      </p>

      {SETTINGS_SECTIONS.map(({ title, description, icon, settings }) => (
        <Card key={title} className="p-4 gap-2">
          <CardTitle className="border-b pb-1">
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
              <SecuritySection
                onPasswordClick={() =>
                  setModals((prev) => ({ ...prev, password: true }))
                }
                onApiKeyClick={() =>
                  setModals((prev) => ({ ...prev, apiKey: true }))
                }
              />
            )}

            {title === "Database" && (
              <DatabaseSection
                loading={loading}
                setLoading={setLoading}
                fileInput={fileInput}
                setFile={setFile}
                setModals={setModals}
              />
            )}

            {title === "Logging" && (
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
            )}

            {title === "Alerts" && (
              <AlertsSection
                onConfigureClick={() =>
                  setModals((prev) => ({ ...prev, notifications: true }))
                }
              />
            )}

            {settings.length > 0 && (
              <DynamicSettingsSection
                settings={settings}
                getSettingValue={getSettingValue}
                handleSettingChange={handleSettingChange}
              />
            )}
          </div>
        </Card>
      ))}

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
