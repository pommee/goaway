"use client";

import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { SettingRow } from "./SettingsRow";
import { Switch } from "@/components/ui/switch";

export function LoggingSection({
  loggingEnabled,
  logLevel,
  statisticsRetention,
  onLoggingEnabledChange,
  onLogLevelChange,
  onStatisticsRetentionChange
}: {
  loggingEnabled?: boolean;
  logLevel?: number;
  statisticsRetention?: number;
  onLoggingEnabledChange?: (enabled: boolean) => void;
  onLogLevelChange?: (level: string) => void;
  onStatisticsRetentionChange?: (days: string) => void;
}) {
  const logLevelMap = ["debug", "info", "warning", "error"];
  const currentLogLevel =
    typeof logLevel === "number" ? logLevelMap[logLevel] : logLevel;

  return (
    <>
      <SettingRow
        title="Log Level"
        description="Controls the verbosity of system logs."
        action={
          <ToggleGroup
            type="single"
            variant="outline"
            value={currentLogLevel || "info"}
            onValueChange={(value) => {
              if (value && onLogLevelChange) {
                onLogLevelChange(value);
              }
            }}
          >
            <ToggleGroupItem value="debug">Debug</ToggleGroupItem>
            <ToggleGroupItem value="info">Info</ToggleGroupItem>
            <ToggleGroupItem value="warning">Warning</ToggleGroupItem>
            <ToggleGroupItem value="error">Error</ToggleGroupItem>
          </ToggleGroup>
        }
      />

      <SettingRow
        title="Statistics Retention"
        description="Number of days to retain system statistics."
        action={
          <ToggleGroup
            type="single"
            variant="outline"
            value={String(statisticsRetention || 7)}
            onValueChange={(value) => {
              if (value && onStatisticsRetentionChange) {
                onStatisticsRetentionChange(value);
              }
            }}
          >
            <ToggleGroupItem value="1">1 day</ToggleGroupItem>
            <ToggleGroupItem value="7">7 days</ToggleGroupItem>
            <ToggleGroupItem value="30">30 days</ToggleGroupItem>
            <ToggleGroupItem value="90">90 days</ToggleGroupItem>
          </ToggleGroup>
        }
      />

      <SettingRow
        title="Logging"
        description="Enable or disable logging across the system."
        action={
          <Switch
            id="logging-enabled"
            checked={loggingEnabled ?? true}
            onCheckedChange={onLoggingEnabledChange}
          />
        }
      />
    </>
  );
}
