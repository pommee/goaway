import { AddList } from "@/app/lists/AddList";
import { ListCard } from "@/app/lists/card";
import { UpdateCustom } from "@/app/lists/updateCustom";
import { DeleteRequest, GetRequest } from "@/util";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type ListEntry = {
  name: string;
  url: string;
  active: boolean;
  blockedCount: number;
  lastUpdated: number;
};

export function Blacklist() {
  const [lists, setLists] = useState<ListEntry[]>([]);
  const [blockedDomains, setBlockedDomains] = useState<number>(0);
  const [editMode, setEditMode] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [updating, setUpdating] = useState<Set<string>>(new Set());
  const [deleting, setDeleting] = useState<Set<string>>(new Set());
  const [fadingOut, setFadingOut] = useState<Set<string>>(new Set());

  const sortLists = (lists: ListEntry[]) => {
    return lists.sort((a, b) => {
      if (a.name === "Custom") return -1;
      if (b.name === "Custom") return 1;
      return a.name.localeCompare(b.name);
    });
  };

  useEffect(() => {
    async function fetchLists() {
      const [code, response] = await GetRequest("lists");
      if (code !== 200) {
        toast.warning("Unable to fetch lists");
        return;
      }

      const listArray: ListEntry[] = Object.entries(response).map(
        ([name, details]) => ({
          name,
          ...details
        })
      );

      const sortedListArray = sortLists(listArray);
      setLists(sortedListArray);

      const totalBlockedDomains = sortedListArray
        .filter((list) => list.active)
        .reduce((total, list) => total + list.blockedCount, 0);

      setBlockedDomains(totalBlockedDomains);
    }

    fetchLists();
  }, []);

  const handleDelete = (name: string, url: string) => {
    setDeleting((prev) => new Set(prev).add(name + url));
    setTimeout(() => {
      setFadingOut((prev) => new Set(prev).add(name + url));
      setTimeout(() => {
        setLists((prevLists) =>
          prevLists.filter((list) => !(list.name === name && list.url === url))
        );
        setDeleting((prev) => {
          const next = new Set(prev);
          next.delete(name + url);
          return next;
        });
        setFadingOut((prev) => {
          const next = new Set(prev);
          next.delete(name + url);
          return next;
        });
      }, 400);
    }, 0);
  };

  const handleRename = (oldName: string, url: string, newName: string) => {
    setLists((prevLists) =>
      prevLists.map((list) =>
        list.name === oldName && list.url === url
          ? { ...list, name: newName }
          : list
      )
    );
    setSelected((prev) => {
      const oldKey = oldName;
      const newKey = newName;
      if (prev.has(oldKey)) {
        const next = new Set(prev);
        next.delete(oldKey);
        next.add(newKey);
        return next;
      }
      return prev;
    });
  };

  const handleListAdded = (newList: ListEntry) => {
    setLists((prev) => sortLists([...prev, newList]));

    if (newList.active) {
      setBlockedDomains((prev) => prev + newList.blockedCount);
    }
  };

  const handleSelect = (name: string, url: string) => {
    setSelected((prev) => {
      const key = `${name}|${url}`;
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const handleRemoveSelected = async () => {
    for (const key of selected) {
      const [name, url] = key.split("|");
      setDeleting((prev) => new Set(prev).add(key));

      await DeleteRequest(
        `list?name=${encodeURIComponent(name)}&url=${encodeURIComponent(url)}`,
        null
      );

      setTimeout(() => {
        setFadingOut((prev) => new Set(prev).add(key));
        setTimeout(() => {
          setLists((prev) =>
            prev.filter((list) => !(list.name === name && list.url === url))
          );

          setDeleting((prev) => {
            const next = new Set(prev);
            next.delete(key);
            return next;
          });
          setFadingOut((prev) => {
            const next = new Set(prev);
            next.delete(key);
            return next;
          });
        }, 400);
      }, 0);
    }

    setSelected(new Set());
  };

  const handleUpdateSelected = async () => {
    let updatedCount = 0;
    const updatingNow = new Set(selected);
    setUpdating(new Set(updatingNow));
    for (const name of selected) {
      const listEntry = lists.find((list) => list.name === name);
      if (!listEntry) {
        updatingNow.delete(name);
        setUpdating(new Set(updatingNow));
        continue;
      }
      const [diffCode, diffResp] = await GetRequest(
        `fetchUpdatedList?name=${encodeURIComponent(listEntry.name)}&url=${
          listEntry.url || ""
        }`
      );
      if (diffCode === 200 && diffResp.updateAvailable) {
        const [code] = await GetRequest(
          `runUpdateList?name=${encodeURIComponent(listEntry.name)}&url=${
            listEntry.url || ""
          }`
        );
        if (code === 200) updatedCount++;
      }
      updatingNow.delete(name);
      setUpdating(new Set(updatingNow));
    }
    toast.info(`${updatedCount} list(s) updated`);
    setEditMode(false);
    setSelected(new Set());
  };

  return (
    <div>
      <div className="flex gap-5 items-center">
        <AddList onListAdded={handleListAdded} />
        <UpdateCustom />
        <div className="flex gap-4 mb-4">
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-accent border-b rounded-t-sm border-b-blue-400">
            <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
            <span className="text-muted-foreground text-sm">Total Lists:</span>
            <span className="font-semibold">{lists.length}</span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-accent border-b rounded-t-sm border-b-green-400">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <span className="text-muted-foreground text-sm">Active:</span>
            <span className="font-semibold">
              {lists.filter((list) => list.active).length}
            </span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-accent border-b rounded-t-sm border-b-red-400">
            <div className="w-2 h-2 bg-red-500 rounded-full"></div>
            <span className="text-muted-foreground text-sm">Inactive:</span>
            <span className="font-semibold">
              {lists.filter((list) => !list.active).length}
            </span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-accent border-b rounded-t-sm border-b-orange-400">
            <div className="w-2 h-2 bg-red-500 rounded-full"></div>
            <span className="text-muted-foreground text-sm">
              Blocked Domains:
            </span>
            <span className="font-semibold">
              {blockedDomains.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
      <div className="flex gap-2 mb-2">
        <Button variant="outline" onClick={() => setEditMode((v) => !v)}>
          {editMode ? "Exit Edit Mode" : "Edit Lists"}
        </Button>
        {editMode && (
          <>
            <Button
              onClick={handleRemoveSelected}
              disabled={selected.size === 0}
              className="bg-red-600 text-white"
            >
              Remove Selected
            </Button>
            <Button
              onClick={handleUpdateSelected}
              disabled={selected.size === 0}
              className="bg-blue-600 text-white"
            >
              Update Selected
            </Button>
          </>
        )}
      </div>
      <div className="grid lg:grid-cols-3 gap-2">
        {lists.map((list, index) => (
          <ListCard
            key={index}
            {...list}
            onDelete={() => handleDelete(list.name, list.url)}
            onRename={handleRename}
            editMode={editMode}
            onSelect={() => handleSelect(list.name, list.url)}
            selected={selected.has(`${list.name}|${list.url}`)}
            updating={updating.has(list.name)}
            deleting={deleting.has(`${list.name}|${list.url}`)}
            fadingOut={fadingOut.has(`${list.name}|${list.url}`)}
          />
        ))}
      </div>
    </div>
  );
}
