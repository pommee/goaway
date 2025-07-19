import { Button } from "@/components/ui/button";
import { IPEntry } from "@/pages/logs";
import { DeleteRequest, PostRequest } from "@/util";
import {
  CaretDownIcon,
  CaretUpIcon,
  CheckIcon,
  LightningIcon,
  ShieldSlashIcon
} from "@phosphor-icons/react";
import { ColumnDef, Header } from "@tanstack/react-table";
import { toast } from "sonner";

type Client = {
  ip: string;
  name: string;
  mac: string;
};

export type Queries = {
  blocked: boolean;
  cached: boolean;
  client: Client;
  domain: string;
  ip: IPEntry[];
  queryType: string;
  responseTimeNS: number;
  status: string;
  timestamp: string;
  protocol: string;
};

interface Props {
  column: Header<object, unknown>["column"];
  title: string;
}

export function SortableHeader({ column, title }: Props) {
  const isSorted = column.getIsSorted();

  const handleClick = () => {
    if (!isSorted) column.toggleSorting(true);
    else if (isSorted === "desc") column.toggleSorting(false);
    else column.clearSorting();
  };

  return (
    <Button
      variant="ghost"
      onClick={handleClick}
      className="hover:text-green-400 hover:bg-stone-900 !p-0 h-auto cursor-pointer gap-1"
    >
      {title}
      {isSorted === "asc" && (
        <CaretUpIcon className="h-4 w-4 text-orange-400" />
      )}
      {isSorted === "desc" && (
        <CaretDownIcon className="h-4 w-4 text-orange-400" />
      )}
    </Button>
  );
}

async function BlacklistDomain(domain: string) {
  try {
    await DeleteRequest(`whitelist?domain=${domain}`, null, true);

    const [status] = await PostRequest("custom", { domains: [domain] });
    if (status === 200) {
      toast.success(`Blacklisted ${domain}`);
    } else {
      toast.error(`Failed to block ${domain}`);
    }
  } catch {
    toast.error("An error occurred while sending the request.");
  }
}

async function WhitelistDomain(domain: string) {
  await DeleteRequest(`blacklist?domain=${domain}`, null, true);

  const [code, response] = await PostRequest("whitelist", {
    domain: domain
  });
  if (code === 200) {
    toast.success(`Whitelisted ${domain}`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

export const columns: ColumnDef<Queries>[] = [
  {
    accessorKey: "timestamp",
    header: ({ column }) => <SortableHeader column={column} title="Time" />,
    cell: ({ row }) => {
      try {
        const timestamp = row.original.timestamp;
        const date = new Date(timestamp).toLocaleString("en-US", {
          month: "short",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
          hour12: false
        });

        return <div className="text-muted-foreground">{date}</div>;
      } catch {
        return (
          <div className="text-muted-foreground">{row.original.timestamp}</div>
        );
      }
    }
  },
  {
    accessorKey: "domain",
    header: ({ column }) => <SortableHeader column={column} title="Domain" />,
    cell: ({ row }) => {
      const wasBlocked = row.original.blocked === true ? "text-red-500" : "";
      return <div className={`${wasBlocked}`}>{row.getValue("domain")}</div>;
    }
  },
  {
    accessorKey: "ip",
    header: "IP(s)",
    cell: ({ getValue }) => {
      const value = getValue() as IPEntry[];
      if (Array.isArray(value)) {
        return (
          <div className="flex flex-col">
            {value.map((entry, i) => {
              if (entry && typeof entry === "object" && entry.ip) {
                const ip = String(entry.ip || "");
                const rtype = String(entry.rtype || "");
                return (
                  <span key={i}>
                    {ip} {rtype && `(${rtype})`}
                  </span>
                );
              } else {
                return <span key={i}>{String(entry || "")}</span>;
              }
            })}
          </div>
        );
      }
      return <div>{String(value || "")}</div>;
    }
  },
  {
    id: "client",
    header: "Client",
    cell: ({ row }) => {
      const client = row.original.client;
      return (
        <div>
          {client.name} | {client.ip}
        </div>
      );
    }
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const query = row.original;
      const wasOK =
        query.blocked === false
          ? query.cached
            ? "cache (forwarded)"
            : "ok (forwarded)"
          : "blacklisted";

      return (
        <div className="flex">
          {query.blocked === false ? (
            query.cached ? (
              <LightningIcon size={14} color="yellow" className="mt-1 mr-1" />
            ) : (
              <CheckIcon size={14} color="green" className="mt-1 mr-1" />
            )
          ) : (
            <ShieldSlashIcon size={14} color="red" className="mt-1 mr-1" />
          )}
          <div className="border-1 px-1 border-stone-800 rounded-sm mr-1">
            {wasOK}
          </div>
          <div>{query.status}</div>
        </div>
      );
    }
  },
  {
    accessorKey: "responseTimeNS",
    header: ({ column }) => <SortableHeader column={column} title="Response" />,
    cell: ({ row }) => {
      const ns = row.original.responseTimeNS;
      const ms = ns / 1_000_000;
      const rowText =
        ms < 10 ? `${Math.round(ns / 1_000)} µs` : `${ms.toFixed(2)} ms`;

      return <p>{rowText}</p>;
    }
  },
  {
    accessorKey: "queryType",
    header: ({ column }) => <SortableHeader column={column} title="Type" />,
    cell: ({ row }) => <div>{row.getValue("queryType")}</div>
  },
  {
    accessorKey: "protocol",
    header: ({ column }) => <SortableHeader column={column} title="Protocol" />,
    cell: ({ row }) => <div>{row.getValue("protocol")}</div>
  },
  {
    accessorKey: "responseSizeBytes",
    header: ({ column }) => <SortableHeader column={column} title="Size" />,
    cell: ({ row }) => <div>{row.getValue("responseSizeBytes")}</div>
  },
  {
    accessorKey: "action",
    header: "Action",
    cell: ({ row }) => {
      const isBlocked = row.original.blocked;

      const handleClick = () => {
        if (isBlocked) {
          WhitelistDomain(row.original.domain);
        } else {
          BlacklistDomain(row.original.domain);
        }
      };

      return (
        <div className="flex justify-center items-center w-fit">
          {isBlocked === false ? (
            <div
              onClick={handleClick}
              className="rounded-sm text-red-500 border-1 px-2 py-0.5  hover:bg-stone-800 transition-colors cursor-pointer text-sm"
            >
              Block
            </div>
          ) : (
            <div
              onClick={handleClick}
              className="rounded-sm text-green-500 border-1 px-2 py-0.5  hover:bg-stone-800 transition-colors cursor-pointer text-sm"
            >
              Allow
            </div>
          )}
        </div>
      );
    }
  }
];
