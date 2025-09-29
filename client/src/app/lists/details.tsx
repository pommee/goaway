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
import { DeleteRequest, GetRequest, PatchRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  EraserIcon,
  EyeIcon,
  ToggleLeftIcon,
  PencilIcon,
  CheckIcon,
  XIcon
} from "@phosphor-icons/react";
import { useState } from "react";
import TimeAgo from "react-timeago";
import { toast } from "sonner";
import BlockedDomainsList from "./blockedDomains";

export function CardDetails(
  listEntry: ListEntry & {
    onDelete?: (name: string) => void;
    onRename?: (name: string, url: string) => void;
  }
) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [fetchingDiff, setFetchingDiff] = useState(false);
  const [deletingList, setDeletingList] = useState(false);
  const [listActive, setListActive] = useState<boolean>(listEntry.active);
  const [updateDiff, setUpdateDiff] = useState({
    diffAdded: [],
    diffRemoved: []
  });
  const [showDiff, setShowDiff] = useState(false);
  const [listUpdating, setListUpdating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editedName, setEditedName] = useState(listEntry.name);
  const [updatingName, setUpdatingName] = useState(false);

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
    setListUpdating(true);
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
        `list?name=${encodeURIComponent(
          listEntry.name
        )}&url=${encodeURIComponent(listEntry.url)}`,
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
    } catch {
      toast.error("Error deleting list");
      setShowDiff(false);
    }

    setDeletingList(false);
  };

  const updateListName = async () => {
    if (editedName.trim() === "") {
      toast.warning("List name cannot be empty");
      setIsEditing(false);
      setEditedName(listEntry.name);
      return;
    }

    if (editedName === listEntry.name) {
      toast.info("List name was not changed");
      setIsEditing(false);
      setEditedName(listEntry.name);
      return;
    }

    setUpdatingName(true);
    try {
      const [code, response] = await PatchRequest(
        `listName?old=${listEntry.name}&new=${editedName.trim()}&url=${
          listEntry.url
        }`
      );

      if (code === 200) {
        toast.success(`List renamed to "${editedName.trim()}"`);
        listEntry.onRename?.(editedName.trim(), listEntry.url);
        setIsEditing(false);
      } else {
        toast.error(response.error || "Failed to update list name");
        setEditedName(listEntry.name);
        setIsEditing(false);
      }
    } catch (error) {
      console.log(error);
      toast.error("Error updating list name");
      setEditedName(listEntry.name);
      setIsEditing(false);
    }

    setUpdatingName(false);
  };

  const cancelEdit = () => {
    setEditedName(listEntry.name);
    setIsEditing(false);
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button
          variant="secondary"
          className="w-full mt-2 rounded-lg text-sm py-1 h-auto"
        >
          <EyeIcon className="mr-2" size={16} />
          View Details
        </Button>
      </DialogTrigger>
      <DialogContent className="rounded-xl w-2/3 max-w-200">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {isEditing ? (
              <div className="flex items-center gap-2 flex-1">
                <input
                  type="text"
                  value={editedName}
                  onChange={(e) => setEditedName(e.target.value)}
                  className="flex-1 px-2 py-1 text-lg font-semibold bg-background border border-border rounded focus:outline-none focus:ring-2 focus:ring-ring"
                  autoFocus
                  disabled={updatingName}
                />
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={updateListName}
                  disabled={updatingName}
                  className="h-8 w-8 p-0 hover:bg-green-100"
                >
                  {updatingName ? (
                    <ArrowsClockwiseIcon className="animate-spin" size={16} />
                  ) : (
                    <CheckIcon className="text-green-600" size={16} />
                  )}
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={cancelEdit}
                  disabled={updatingName}
                  className="h-8 w-8 p-0 hover:bg-red-100"
                >
                  <XIcon className="text-red-600" size={16} />
                </Button>
              </div>
            ) : (
              <>
                <span>{listEntry.name}</span>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => setIsEditing(true)}
                  className="h-8 w-8 p-0 hover:bg-accent"
                >
                  <PencilIcon className="text-muted-foreground" size={16} />
                </Button>
              </>
            )}
          </DialogTitle>
          {listEntry.url && (
            <a
              href={listEntry.url}
              target={"_"}
              className="text-sm text-muted-foreground hover:text-accent-foreground"
            >
              {listEntry.url}
            </a>
          )}
        </DialogHeader>
        <Separator />
        <div>
          <div className="flex gap-1">
            <p className="text-muted-foreground">status:</p>
            {listActive ? (
              <p className="w-fit text-green-500">active</p>
            ) : (
              <p className="w-fit text-red-500">inactive</p>
            )}
          </div>
          <div className="flex gap-1">
            <p className="text-muted-foreground">blocked:</p>
            <p>{listEntry.blockedCount.toLocaleString()}</p>
          </div>
          <div className="flex gap-1">
            <p className="text-muted-foreground">updated:</p>
            <div>
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
            className="bg-green-600 border-none hover:bg-green-500 text-white flex-1 text-sm"
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
                    Looking for update...
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
                variant="destructive"
                className="flex-1"
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
          <div className="mt-4 p-4 bg-accent rounded-lg border">
            <h3 className="font-bold mb-2">
              Update for {listEntry.name} found
            </h3>
            <Separator className="mb-1" />
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
                        className="text-green-400 bg-background p-1 rounded-sm"
                      >
                        {item}
                      </div>
                    ))}
                  </div>
                </div>
                <Separator className="mb-2" />
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
                      className="text-red-400 bg-background p-1 rounded-sm"
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
              {listUpdating ? (
                <p className="flex items-center">
                  <ArrowsClockwiseIcon
                    className="mr-2 animate-spin"
                    size={16}
                  />
                  Updating list...
                </p>
              ) : (
                <p className="flex items-center">Accept changes (?)</p>
              )}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
