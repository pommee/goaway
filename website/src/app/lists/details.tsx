import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { ListEntry } from "@/pages/lists";
import { GetRequest } from "@/util";
import {
  ArrowsClockwise,
  Eraser,
  Eye,
  ToggleLeft
} from "@phosphor-icons/react";
import { useState } from "react";
import { toast } from "sonner";
import TimeAgo from "react-timeago";
import { Separator } from "@/components/ui/separator";
import BlockedDomainsList from "./blockedDomains";

export function CardDetails(listEntry: ListEntry) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const toggleBlocklist = async () => {
    const [code, response] = await GetRequest(
      `toggleBlocklist?blocklist=${listEntry.name}`
    );
    if (code === 200) {
      toast.info(response.message);
      setDialogOpen(false);
    } else {
      toast.error(response.message);
      setDialogOpen(false);
    }
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="bg-zinc-800 border-none hover:bg-zinc-700 text-white w-full mt-2 rounded-lg text-sm py-1 h-auto"
        >
          <Eye className="mr-2" size={16} />
          View Details
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-zinc-900 text-white border-zinc-800 rounded-xl w-2/3 max-w-200">
        <DialogHeader>
          <DialogTitle>{listEntry.name} </DialogTitle>
          {listEntry.url && (
            <a
              href={listEntry.url}
              className="text-sm text-zinc-500 hover:text-zinc-300"
            >
              {listEntry.url}
            </a>
          )}
        </DialogHeader>
        <Separator className="bg-zinc-800" />
        <div>
          <div className="flex gap-1 text-zinc-500">
            {"status:"}
            {listEntry.active == true ? (
              <p className="w-fit text-green-500">active</p>
            ) : (
              <p className="w-fit text-red-500">inactive</p>
            )}
          </div>
          <div className="flex gap-1 text-zinc-500">
            {"blocked:"}
            <p className="text-white">
              {listEntry.blockedCount.toLocaleString()}
            </p>
          </div>
          <div className="flex gap-1 text-zinc-500">
            {"updated:"}
            <div className="text-white">
              <TimeAgo
                date={new Date(listEntry.lastUpdated * 1000)}
                minPeriod={60}
              />
            </div>
          </div>
        </div>
        <div className="flex gap-2 justify-between flex-wrap">
          <Button
            onClick={toggleBlocklist}
            variant="outline"
            className="bg-zinc-800 border-none hover:bg-zinc-700 text-white flex-1 text-sm"
          >
            <ToggleLeft className="mr-1" size={16} />
            Toggle
          </Button>

          {listEntry.name !== "Custom" && (
            <>
              <Button
                variant="outline"
                className="bg-blue-600 border-none hover:bg-blue-500 text-white flex-1 text-sm"
              >
                <ArrowsClockwise className="mr-1" size={16} />
                Update [WIP]
              </Button>
              <Button
                variant="outline"
                className="bg-red-600 border-none hover:bg-red-500 text-white flex-1 text-sm"
              >
                <Eraser className="mr-1" size={16} />
                Delete
              </Button>
            </>
          )}
        </div>
        {listEntry.name === "Custom" && (
          <BlockedDomainsList listName={listEntry.name} />
        )}
      </DialogContent>
    </Dialog>
  );
}
