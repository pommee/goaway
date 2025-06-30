import { AddUpstream } from "@/app/upstream/addUpstream";
import { UpstreamCard } from "@/app/upstream/card";
import { GetRequest } from "@/util";
import { ArrowClockwiseIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type UpstreamEntry = {
  dnsPing: string;
  icmpPing: string;
  name: string;
  preferred: boolean;
  upstream: string;
};

export function Upstream() {
  const [upstreams, setUpstreams] = useState<UpstreamEntry[]>([]);

  const fetchupstreams = async () => {
    const [code, response] = await GetRequest("upstreams");
    if (code !== 200) {
      toast.warning(`Unable to fetch upstreams`);
      return;
    }
    setUpstreams(response.upstreams);
  };

  useEffect(() => {
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
        <div className="grid lg:grid-cols-4 gap-2">
          {upstreams.map((upstream) => (
            <UpstreamCard
              key={upstream.upstream}
              upstream={upstream}
              onRemove={handleRemoveUpstream}
            />
          ))}
        </div>
      )) || (
        <div className="flex justify-center items-center">
          <div className="flex flex-col items-center space-y-4">
            <ArrowClockwiseIcon className="w-12 h-12 text-blue-400 animate-spin" />
            <p className="text-lg text-stone-400">Loading upstreams...</p>
          </div>
        </div>
      )}
    </div>
  );
}
