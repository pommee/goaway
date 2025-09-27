import { Button } from "@/components/ui/button";
import { IPEntry } from "@/pages/logs";
import { CaretDownIcon, CaretUpIcon } from "@phosphor-icons/react";
import { Header } from "@tanstack/react-table";

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

function something() {
  return "awd";
}
