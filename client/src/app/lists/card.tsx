import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { ListEntry } from "@/pages/blacklist";
import {
  ArrowUpIcon,
  ClockIcon,
  ShieldSlashIcon,
  EraserIcon
} from "@phosphor-icons/react";
import { CardDetails } from "./details";
import { Checkbox } from "@/components/ui/checkbox";

export function ListCard(
  props: ListEntry & {
    onDelete: (name: string, url: string) => void;
    onRename: (oldName: string, url: string, newName: string) => void;
    editMode?: boolean;
    selected?: boolean;
    onSelect?: () => void;
    updating?: boolean;
    deleting?: boolean;
    fadingOut?: boolean;
  }
) {
  const {
    editMode,
    selected,
    onSelect,
    updating,
    deleting,
    fadingOut,
    ...listEntry
  } = props;

  const formattedDate = new Date(listEntry.lastUpdated * 1000).toLocaleString(
    "en-US",
    {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false
    }
  );

  return (
    <Card
      className={`w-full p-6 rounded-2xl relative shadow-lg hover:shadow-xl transition-all duration-300 border
        ${
          fadingOut
            ? "opacity-0 transition-opacity duration-400"
            : "opacity-100"
        }
      `}
    >
      {editMode && listEntry.name !== "Custom" && (
        <div className="absolute top-4 left-4 z-10 cursor-pointer">
          <Checkbox checked={selected} onCheckedChange={onSelect} />
        </div>
      )}
      {deleting && (
        <div className="absolute top-4 right-10 z-20">
          <EraserIcon className="animate-bounce text-red-500" size={22} />
        </div>
      )}
      <div
        className={`absolute top-4 right-4 w-2 h-2 rounded-full ${
          listEntry.active ? "bg-green-500" : "bg-red-500"
        } shadow-glow`}
      />
      <div className="flex flex-col gap-4">
        <div className="w-full">
          <h2 className="text-center text-xl font-bold mb-1">
            <p className="flex items-center justify-center gap-2">
              {listEntry.name}{" "}
              {updating && (
                <ArrowUpIcon
                  className="animate-bounce text-green-500"
                  size={22}
                />
              )}
            </p>
          </h2>
          <p className="text-center text-xs text-muted-foreground truncate">
            {listEntry.url}
          </p>
          <Separator className="mt-1" />
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center bg-accent rounded-full px-3 py-1 text-sm">
            <ShieldSlashIcon className="mr-1" size={14} />
            <span>{listEntry.blockedCount}</span>
          </div>

          <div className="flex items-center text-muted-foreground text-sm">
            <ClockIcon className="mr-1" size={14} />
            <span>{formattedDate}</span>
          </div>
        </div>

        <CardDetails
          {...listEntry}
          onDelete={() => props.onDelete(listEntry.name, listEntry.url)}
          onRename={(newName: string, url: string) =>
            props.onRename(listEntry.name, url, newName)
          }
        />
      </div>
    </Card>
  );
}
