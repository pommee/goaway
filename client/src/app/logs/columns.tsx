import { IPEntry } from "@/pages/logs";
import { Check, Lightning, ShieldSlash } from "@phosphor-icons/react";
import { Checkbox } from "@radix-ui/react-checkbox";
import { ColumnDef } from "@tanstack/react-table";

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
};

export const columns: ColumnDef<Queries>[] = [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false
  },
  {
    accessorKey: "timestamp",
    header: "Time",
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
        return <div>{row.original.timestamp}</div>;
      }
    }
  },
  {
    accessorKey: "domain",
    header: "Domain",
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
            ? `cache (forwarded)`
            : `ok (forwarded)`
          : `blacklisted`;
      const ns = query.responseTimeNS;
      const ms = ns / 1_000_000;
      const rowText =
        ms < 10 ? `${Math.round(ns / 1_000)}µs` : `${ms.toFixed(2)}ms`;

      return (
        <div className="flex">
          {query.blocked === false ? (
            query.cached ? (
              <Lightning size={14} color="yellow" className="mt-1 mr-1" />
            ) : (
              <Check size={14} color="green" className="mt-1 mr-1" />
            )
          ) : (
            <ShieldSlash size={14} color="red" className="mt-1 mr-1" />
          )}
          <div className="border-1 px-1 border-stone-800 rounded-sm mr-1">
            {wasOK}
          </div>
          <div>
            {query.status} {rowText}
          </div>
        </div>
      );
    }
  },
  {
    accessorKey: "queryType",
    header: "Type",
    cell: ({ row }) => <div>{row.getValue("queryType")}</div>
  },
  {
    accessorKey: "responseSizeBytes",
    header: "Size",
    cell: ({ row }) => <div>{row.getValue("responseSizeBytes")}</div>
  }
];
