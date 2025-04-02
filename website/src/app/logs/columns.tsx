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
  ip: Array<string>;
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
        const date = new Date(timestamp);

        if (isNaN(date.getTime())) {
          return <div>{timestamp}</div>;
        }

        const formattedDate = `${date.getFullYear()}/${String(
          date.getMonth() + 1
        ).padStart(2, "0")}/${String(date.getDate()).padStart(2, "0")} ${String(
          date.getHours()
        ).padStart(2, "0")}:${String(date.getMinutes()).padStart(
          2,
          "0"
        )}:${String(date.getSeconds()).padStart(2, "0")}`;

        return <div>{formattedDate}</div>;
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
    header: "IP",
    cell: ({ row }) => <div>{row.getValue("ip")}</div>
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
            ? `cache (forwarded) ${query.status}`
            : `ok (forwarded) ${query.status}`
          : query.status;
      const responseTimeMS = (query.responseTimeNS / 1_000_000).toFixed(2);
      const rowText = ` ${wasOK} ${responseTimeMS}ms`;
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
          {rowText}
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
