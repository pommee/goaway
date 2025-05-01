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
  const [updateDiff, setUpdateDiff] = useState({
    diffAdded: [],
    diffRemoved: []
  });
  const [showDiff, setShowDiff] = useState(false);

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

  const checkForUpdates = async () => {
    try {
      const [code, response] = await GetRequest(
        `fetchUpdatedList?name=${encodeURIComponent(listEntry.name)}&url=${
          listEntry.url || ""
        }`
      );

      if (code === 200) {
        if (response.updateAvailable) {
          setUpdateDiff({
            diffAdded: response.diffAdded || [],
            diffRemoved: response.diffRemoved || []
          });
          setShowDiff(true);
        } else {
          toast.info("No updates available");
          setShowDiff(false);
        }
      } else {
        toast.error(response.error);
        setShowDiff(false);
      }
    } catch (error) {
      console.log(error);
      toast.error("Error checking for updates");
      setShowDiff(false);
    }
  };

  const runUpdateList = async () => {
    try {
      const [code, response] = await GetRequest(
        `runUpdateList?name=${encodeURIComponent(listEntry.name)}&url=${
          listEntry.url || ""
        }`
      );

      if (code === 200) {
        setDialogOpen(false);
        setShowDiff(false);
        toast.info(`Updated ${listEntry.name}`);
      } else {
        toast.error(response.error);
        setShowDiff(false);
      }
    } catch (error) {
      console.log(error);
      toast.error("Error checking for updates");
      setShowDiff(false);
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
                onClick={checkForUpdates}
                variant="outline"
                className="bg-blue-600 border-none hover:bg-blue-500 text-white flex-1 text-sm"
              >
                <ArrowsClockwise className="mr-1" size={16} />
                Update
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

        {showDiff && (
          <div className="mt-4 p-4 bg-zinc-800 rounded-lg border border-zinc-700">
            <h3 className="font-bold mb-2">
              Update for {listEntry.name} found
            </h3>
            <Separator className="bg-stone-700 mb-1" />
            {updateDiff.diffAdded.length > 0 && (
              <>
                <div className="mb-3">
                  <h4 className="text-green-400 mb-1">
                    New Domains: {updateDiff.diffAdded.length}{" "}
                  </h4>
                  <div className="max-h-24 overflow-y-auto text-xs flex gap-1 flex-wrap">
                    {updateDiff.diffAdded.map((item, i) => (
                      <div
                        key={`added-${i}`}
                        className="text-green-300 bg-stone-900 p-1 rounded-sm"
                      >
                        {item}
                      </div>
                    ))}
                  </div>
                </div>
                <Separator className="bg-stone-600 mb-2" />
              </>
            )}
            {updateDiff.diffRemoved.length > 0 && (
              <div>
                <h4 className="text-red-400 mb-1">
                  Deleted Domains: {updateDiff.diffRemoved.length}
                </h4>
                <div className="max-h-24 overflow-y-auto text-xs flex gap-1 flex-wrap">
                  {updateDiff.diffRemoved.map((item, i) => (
                    <div
                      key={`removed-${i}`}
                      className="text-red-300 bg-stone-900 p-1 rounded-sm"
                    >
                      {item}
                    </div>
                  ))}
                </div>
              </div>
            )}
            <Button
              className="mt-6 bg-green-700 hover:bg-green-600 cursor-pointer text-white w-full"
              onClick={runUpdateList}
            >
              Accept changes
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
