export interface Root {
  dns: Dns;
  api: Api;
  logging: Logging;
  misc: Misc;
}

export interface Dns {
  status: Status;
  address: string;
  gateway: string;
  cacheTTL: number;
  udpSize: number;
  tls: Tls;
  upstream: Upstream;
  ports: Ports;
}

export interface Status {
  pausedAt: string;
  pauseTime: string;
  paused: boolean;
}

export interface Tls {
  enabled: boolean;
  cert: string;
  key: string;
}

export interface Upstream {
  preferred: string;
  fallback: string[];
}

export interface Ports {
  udptcp: number;
  dot: number;
  doh: number;
}

export interface Api {
  port: number;
  authentication: boolean;
  rateLimit: RateLimit;
}

export interface RateLimit {
  enabled: boolean;
  maxTries: number;
  window: number;
}

export interface Logging {
  enabled: boolean;
  level: number;
}

export interface Misc {
  inAppUpdate: boolean;
  statisticsRetention: number;
  dashboard: boolean;
  scheduledBlacklistUpdates: boolean;
}

export interface SetModalsType {
  password: false;
  apiKey: false;
  importConfirm: false;
  notifications: false;
}
