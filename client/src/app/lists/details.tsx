import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";
import { ListEntry } from "@/pages/blacklist";
import { DeleteRequest, GetRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  EraserIcon,
  EyeIcon,
  ToggleLeftIcon
} from "@phosphor-icons/react";
import { useState } from "react";
import TimeAgo from "react-timeago";
import { toast } from "sonner";
import BlockedDomainsList from "./blockedDomains";

export function CardDetails(
  listEntry: ListEntry & { onDelete?: (name: string) => void }
) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [fetchingDiff, setFetchingDiff] = useState(false);
  const [deletingList, setDeletingList] = useState(false);
  const [listActive, setListActive] = useState<boolean>();
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
      setListActive(!listActive);
      toast.info(response.message);
    } else {
      toast.error(response.message);
    }
  };

  const checkForUpdates = async () => {
    setFetchingDiff(true);
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

    setFetchingDiff(false);
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

  const deleteList = async () => {
    setDeletingList(true);
    try {
      const [code, response] = await DeleteRequest(
        `list?name=${encodeURIComponent(listEntry.name)}`,
        null
      );

      if (code === 200) {
        toast.info("Info", {
          description: `Deleted list: ${listEntry.name}`
        });
        setDialogOpen(false);
        setShowDiff(false);
        listEntry.onDelete?.(listEntry.name);
      } else {
        toast.error(response.error);
        setShowDiff(false);
      }
    } catch (error) {
      console.log(error);
      toast.error("Error deleting list");
      setShowDiff(false);
    }

    setDeletingList(false);
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="bg-zinc-800 border-none hover:bg-zinc-700 text-white w-full mt-2 rounded-lg text-sm py-1 h-auto"
        >
          <EyeIcon className="mr-2" size={16} />
          View Details
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-zinc-900 text-white border-zinc-800 rounded-xl w-2/3 max-w-200">
        <DialogHeader>
          <DialogTitle>{listEntry.name} </DialogTitle>
          {listEntry.url && (
            <a
              href={listEntry.url}
              target={"_"}
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
            {listActive ? (
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
            <ToggleLeftIcon className="mr-1" size={16} />
            Toggle
          </Button>

          {listEntry.name !== "Custom" && (
            <>
              <Button
                disabled={fetchingDiff}
                onClick={checkForUpdates}
                variant="outline"
                className="bg-blue-600 border-none hover:bg-blue-500 text-white flex-1 text-sm"
              >
                {fetchingDiff ? (
                  <div className="flex">
                    <ArrowsClockwiseIcon
                      className="mr-1 mt-0.5 animate-spin"
                      size={16}
                    />
                    Updating...
                  </div>
                ) : (
                  <div className="flex">
                    <ArrowsClockwiseIcon className="mr-1 mt-0.5" size={16} />
                    Update
                  </div>
                )}
              </Button>
              <Button
                disabled={deletingList}
                onClick={deleteList}
                variant="outline"
                className="bg-red-600 border-none hover:bg-red-500 text-white flex-1 text-sm"
              >
                {deletingList ? (
                  <div className="flex">
                    <EraserIcon
                      className="mr-1 mt-0.5 animate-bounce "
                      size={16}
                    />
                    Delete
                  </div>
                ) : (
                  <div className="flex">
                    <EraserIcon className="mr-1 mt-0.5" size={16} />
                    Delete
                  </div>
                )}
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
