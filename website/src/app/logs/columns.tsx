import { Checkbox } from "@radix-ui/react-checkbox";
import { ColumnDef } from "@tanstack/react-table";
import { BanIcon, Verified } from "lucide-react";

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
    enableHiding: false,
  },
  {
    accessorKey: "timestamp",
    header: "Time",
    cell: ({ row }) => {
      const date = new Date(row.original.timestamp);
      const formattedDate = `${date.getFullYear()}/${String(
        date.getMonth() + 1
      ).padStart(2, "0")}/${String(date.getDate()).padStart(2, "0")} ${String(
        date.getHours() + 1
      ).padStart(2, "0")}:${String(date.getMinutes()).padStart(
        2,
        "0"
      )}:${String(date.getSeconds() - 13).padStart(2, "0")}`;
      return <div>{formattedDate}</div>;
    },
  },
  {
    accessorKey: "domain",
    header: "Domain",
    cell: ({ row }) => {
      const wasBlocked = row.original.blocked === true ? "text-red-500" : "";
      return <div className={`${wasBlocked}`}>{row.getValue("domain")}</div>;
    },
  },
  {
    accessorKey: "ip",
    header: "IP",
    cell: ({ row }) => <div>{row.getValue("ip")}</div>,
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
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const query = row.original;
      const wasOK =
        query.blocked == false
          ? `OK (forwarded) ${query.status}`
          : query.status;
      const responseTimeMS = (query.responseTimeNS / 1_000_000).toFixed(2);
      const rowText = ` ${wasOK} | ${responseTimeMS}ms`;
      return (
        <div className="flex">
          {query.blocked === false ? (
            <Verified size={14} color="green" className="mt-1 mr-1" />
          ) : (
            <BanIcon size={14} color="red" className="mt-1 mr-1" />
          )}
          {rowText}
        </div>
      );
    },
  },
  {
    accessorKey: "queryType",
    header: "Type",
    cell: ({ row }) => <div>{row.getValue("queryType")}</div>,
  },
  {
    accessorKey: "responseSizeBytes",
    header: "Size",
    cell: ({ row }) => <div>{row.getValue("responseSizeBytes")}</div>,
  },
];
