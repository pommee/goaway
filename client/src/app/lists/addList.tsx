import { useState } from "react";
import { toast } from "sonner";
import { PlusIcon, SpinnerGapIcon } from "@phosphor-icons/react";

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
import { DialogDescription } from "@radix-ui/react-dialog";

import { ListEntry } from "@/pages/blacklist";
import { GetRequest } from "@/util";

export function AddList({
  onListAdded
}: {
  onListAdded: (list: ListEntry) => void;
}) {
  const [listName, setListName] = useState("");
  const [url, setUrl] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async () => {
    setIsSaving(true);

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
            Add List
          </Button>
        </DialogTrigger>

        <DialogContent className="bg-zinc-900 text-white border-zinc-800 w-full max-w-xl rounded-xl p-6">
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold">
              Add a New List
            </DialogTitle>
          </DialogHeader>

          <DialogDescription className="text-sm text-zinc-400 mt-1">
            Predefined lists can be imported using a name and a URL. Here are a
            few useful sources:
            <ul className="list-disc pl-6 mt-2 space-y-1">
              {[
                {
                  name: "StevenBlack's hosts",
                  url: "https://github.com/StevenBlack/hosts"
                },
                {
                  name: "The Block List Project",
                  url: "https://blocklistproject.github.io/Lists/"
                },
                { name: "FilterLists", url: "https://filterlists.com/" },
                { name: "The Firebog", url: "https://firebog.net/" }
              ].map(({ name, url }) => (
                <li key={url}>
                  <a
                    href={url}
                    target="_blank"
                    className="text-blue-400 hover:underline"
                  >
                    {name}
                  </a>
                </li>
              ))}
            </ul>
          </DialogDescription>

          <div className="space-y-4 mt-6">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label
                htmlFor="name"
                className="text-right text-sm text-zinc-300"
              >
                List Name
              </Label>
              <Input
                id="name"
                value={listName}
                placeholder="e.g. My Blocklist"
                onChange={(e) => setListName(e.target.value)}
                className="col-span-3 bg-zinc-800 border-zinc-700 text-white placeholder-zinc-500"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="url" className="text-right text-sm text-zinc-300">
                URL
              </Label>
              <Input
                id="url"
                value={url}
                placeholder="e.g. https://example.com/list.txt"
                onChange={(e) => setUrl(e.target.value)}
                className="col-span-3 bg-zinc-800 border-zinc-700 text-white placeholder-zinc-500"
              />
            </div>
          </div>

          <div className="mt-6 flex justify-end">
            <Button
              onClick={handleSave}
              disabled={isSaving || !listName || !url}
              className="bg-green-700 hover:bg-green-600 text-white px-6"
            >
              {isSaving ? (
                <span className="flex items-center">
                  <SpinnerGapIcon className="animate-spin mr-2" />
                  Saving...
                </span>
              ) : (
                "Save"
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
