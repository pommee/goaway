import { AddList } from "@/app/lists/addList";
import { ListCard } from "@/app/lists/card";
import { UpdateCustom } from "@/app/lists/updateCustom";
import { GetRequest } from "@/util";
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

  useEffect(() => {
    async function fetchLists() {
      const [code, response] = await GetRequest(`lists`);
      if (code !== 200) {
        toast.warning(`Unable to fetch lists`);
        return;
      }

      const listArray: ListEntry[] = Object.entries(response.lists).map(
        ([name, details]) => ({
          name,
          ...details
        })
      );

      setLists(listArray);

      const totalBlockedDomains = listArray
        .filter((list) => list.active)
        .reduce((total, list) => total + list.blockedCount, 0);

      setBlockedDomains(totalBlockedDomains);
    }

    fetchLists();
  }, []);

  const handleDelete = (name: string) => {
    setLists((prevLists) => prevLists.filter((list) => list.name !== name));
  };

  const handleListAdded = (newList: ListEntry) => {
    setLists((prev) => [...prev, newList]);

    if (newList.active) {
      setBlockedDomains((prev) => prev + newList.blockedCount);
    }
  };

  return (
    <div>
      <div className="flex gap-5 items-center">
        <AddList onListAdded={handleListAdded} />
        <UpdateCustom />
        <div className="flex gap-4 mb-4">
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-zinc-800 border rounded-t-sm border-b-blue-400">
            <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
            <span className="text-zinc-400 text-sm">Total Lists:</span>
            <span className="text-white font-semibold">{lists.length}</span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-zinc-800 border rounded-t-sm border-b-green-400">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <span className="text-zinc-400 text-sm">Active:</span>
            <span className="text-white font-semibold">
              {lists.filter((list) => list.active).length}
            </span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-zinc-800 border rounded-t-sm border-b-red-400">
            <div className="w-2 h-2 bg-red-500 rounded-full"></div>
            <span className="text-zinc-400 text-sm">Inactive:</span>
            <span className="text-white font-semibold">
              {lists.filter((list) => !list.active).length}
            </span>
          </div>
          <div className="flex items-center gap-2 px-4 py-1 mb-1 bg-zinc-800 border rounded-t-sm border-b-orange-400">
            <div className="w-2 h-2 bg-red-500 rounded-full"></div>
            <span className="text-zinc-400 text-sm">Blocked Domains:</span>
            <span className="text-white font-semibold">
              {blockedDomains.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
      <div className="grid lg:grid-cols-3 gap-2">
        {lists.map((list, index) => (
          <ListCard key={index} {...list} onDelete={handleDelete} />
        ))}
      </div>
    </div>
  );
}
