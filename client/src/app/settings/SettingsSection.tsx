import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
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

export const SETTINGS_SECTIONS = [
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
    settings: []
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
