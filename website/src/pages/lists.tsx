import { AddList } from "@/app/lists/addList";
import { ListCard } from "@/app/lists/card";
import { UpdateCustom } from "@/app/lists/updateCustom";
import { GetRequest } from "@/util";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type ListEntry = {
  name: string;
  active: boolean;
  blockedCount: number;
  lastUpdated: number;
};

export function Lists() {
  const [lists, setLists] = useState<ListEntry[]>([]);

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
    }

    fetchLists();
  }, []);

  return (
    <div>
      <div className="flex gap-5">
        <AddList />
        <UpdateCustom />
      </div>
      <div className="grid lg:grid-cols-3 gap-2">
        {lists.map((list, index) => (
          <ListCard
            key={index}
            active={list.active}
            blockedCount={list.blockedCount}
            lastUpdated={list.lastUpdated}
            name={list.name}
          />
        ))}
      </div>
    </div>
  );
}
