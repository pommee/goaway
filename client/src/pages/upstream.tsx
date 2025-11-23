import { AddUpstream } from "@/app/upstream/addUpstream";
import { UpstreamCard } from "@/app/upstream/card";
import { Skeleton } from "@/components/ui/skeleton";
import { GetRequest } from "@/util";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type UpstreamEntry = {
  upstreamName: string;
  dnsPing: string;
  icmpPing: string;
  name: string;
  preferred: boolean;
  upstream: string;
};

export function Upstream() {
  const [upstreams, setUpstreams] = useState<UpstreamEntry[]>([]);

  useEffect(() => {
    const fetchupstreams = async () => {
      const [code, response] = await GetRequest("upstreams");
      if (code !== 200) {
        toast.warning("Unable to fetch upstreams");
        return;
      }
      if (response && typeof response === 'object' && 'upstreams' in response && Array.isArray(response.upstreams)) {
        setUpstreams(response.upstreams as UpstreamEntry[]);
      } else {
        toast.warning('Invalid upstreams data format');
      }
    };

    fetchupstreams();
  }, []);

  const handleAddUpstream = (entry: UpstreamEntry) => {
    setUpstreams((prev) => [...prev, entry]);
  };

  const handleRemoveUpstream = (upstream: string) => {
    setUpstreams((prev) => {
      const filtered = prev.filter((u) => u.upstream !== upstream);
      return filtered;
    });
  };

  return (
    <div>
      <div className="flex gap-5">
        <AddUpstream onAdd={handleAddUpstream} />
      </div>
      {(upstreams.length > 0 && (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4 mt-6">
          {upstreams.map((upstream) => (
            <UpstreamCard
              key={upstream.upstream}
              upstream={upstream}
              onRemove={handleRemoveUpstream}
            />
          ))}
        </div>
      )) || (
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4 mt-6">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      )}
    </div>
  );
}

function SkeletonCard() {
  return (
    <div className="flex flex-col space-y-3">
      <Skeleton className="h-[200px] w-[380px] rounded-xl" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-[200px]" />
      </div>
    </div>
  );
}
