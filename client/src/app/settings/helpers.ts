export const parseLogLevel = (level: number | string): number | string => {
  const levels = ["Debug", "Info", "Warning", "Error"];
  return typeof level === "number" ? levels[level] : levels.indexOf(level);
};

export const formatBytes = (bytes: number): string => {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
};

export const formatDate = (timestamp: number): string =>
  new Date(timestamp).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false
  });
