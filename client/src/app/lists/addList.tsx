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
import { GetRequest } from "@/util";
import { Plus } from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { toast } from "sonner";

async function CreateNewList(listName: string, url: string) {
  const [code] = await GetRequest(`addList?name=${listName}&url=${url}`);
  if (code === 200) {
    toast.success(`${listName} has been added!`);
  }
}

export function AddList() {
  const [listName, setListName] = useState("");
  const [url, setUrl] = useState("");

  const handleSave = () => {
    CreateNewList(listName, url);
  };

  return (
    <div className="mb-5">
      <Dialog>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className="bg-zinc-800 border-none hover:bg-zinc-700 text-white"
          >
            <Plus className="mr-2" size={20} />
            Add list
          </Button>
        </DialogTrigger>
        <DialogContent className="bg-zinc-900 text-white border-zinc-800 w-2/3 max-w-none">
          <DialogHeader>
            <DialogTitle>New List</DialogTitle>
          </DialogHeader>
          <DialogDescription className="text-base leading-relaxed">
            Predefined lists can be imported, all you need is a name and url.
            You can find lists from various sources online. Some popular sources
            are:
            <ul className="list-disc pl-6 mt-2 space-y-2">
              <li className="text-gray-600">
                <a
                  href="https://github.com/StevenBlack/hosts"
                  className="text-blue-500 hover:underline"
                  target="_blank"
                >
                  StevenBlack's hosts
                </a>
              </li>
              <li className="text-gray-600">
                <a
                  href="https://blocklistproject.github.io/Lists/"
                  className="text-blue-500 hover:underline"
                  target="_blank"
                >
                  The Block List Project
                </a>
              </li>
              <li className="text-gray-600">
                <a
                  href="https://filterlists.com/"
                  className="text-blue-500 hover:underline"
                  target="_blank"
                >
                  FilterLists
                </a>
              </li>
              <li className="text-gray-600">
                <a
                  href="https://firebog.net/"
                  className="text-blue-500 hover:underline"
                  target="_blank"
                >
                  The Firebog
                </a>
              </li>
            </ul>
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
          </div>{" "}
          <Button
            variant="outline"
            className="bg-green-800 border-none hover:bg-green-600 text-white"
            onClick={handleSave}
          >
            Save
          </Button>
        </DialogContent>
      </Dialog>
    </div>
  );
}
