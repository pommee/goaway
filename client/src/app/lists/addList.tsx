import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ListEntry } from "@/pages/blacklist";
import { GetRequest } from "@/util";
import { PlusIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { toast } from "sonner";

export function AddList({
  onListAdded
}: {
  onListAdded: (list: ListEntry) => void;
}) {
  const [listName, setListName] = useState("");
  const [url, setUrl] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  async function CreateNewList(listName: string, url: string) {
    const [code, response] = await GetRequest(
      `addList?name=${listName}&url=${url}`
    );
    if (code === 200) {
      const newList: ListEntry = {
        name: listName,
        url,
        active: response.list.active,
        blockedCount: response.list.blockedCount,
        lastUpdated: response.list.lastUpdated
      };

      onListAdded(newList);
      toast.success(`${listName} has been added!`);
      setModalOpen(false);
    }

    setIsSaving(false);
    setListName("");
    setUrl("");
  }

  const handleSave = () => {
    setIsSaving(true);
    CreateNewList(listName, url);
  };

  return (
    <div className="mb-5">
      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className="bg-zinc-800 border-none hover:bg-zinc-700 text-white"
          >
            <PlusIcon className="mr-2" size={20} />
            Add list
          </Button>
        </DialogTrigger>
        <DialogContent className="bg-zinc-900 text-white border-zinc-800 w-2/4 max-w-none">
          <DialogHeader>
            <DialogTitle>New List</DialogTitle>
          </DialogHeader>
          <DialogDescription className="text-base leading-relaxed">
            Predefined lists can be imported, all you need is a name and url.
            You can find lists from various sources online...
          </DialogDescription>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="name" className="text-right">
                List name
              </Label>
              <Input
                id="name"
                value={listName}
                placeholder="My blocklist"
                onChange={(e) => setListName(e.target.value)}
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="url" className="text-right">
                URL
              </Label>
              <Input
                id="url"
                value={url}
                placeholder="https://example/list"
                onChange={(e) => setUrl(e.target.value)}
                className="col-span-3"
              />
            </div>
          </div>
          <Button
            disabled={isSaving}
            variant="outline"
            className="bg-green-800 border-none hover:bg-green-600 text-white"
            onClick={handleSave}
          >
            {isSaving ? (
              <div className="flex">
                <SpinnerGapIcon className="mt-0.5 mr-2 animate-spin" />
                Saving...
              </div>
            ) : (
              <>Save</>
            )}
          </Button>
        </DialogContent>
      </Dialog>
    </div>
  );
}
