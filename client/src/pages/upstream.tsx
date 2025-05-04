import { AddUpstream } from "@/app/upstream/addUpstream";
import { UpstreamCard } from "@/app/upstream/card";
import { GetRequest } from "@/util";
import { ArrowClockwise } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

type UpstreamResponse = {
  preferredUpstream: string;
  upstreams: [UpstreamEntry];
};

export type UpstreamEntry = {
  dnsPing: string;
  icmpPing: string;
  name: string;
  preferred: boolean;
  upstream: string;
};

export function Upstream() {
  const [upstreams, setUpstreams] = useState<UpstreamResponse>();

  useEffect(() => {
    async function fetchupstreams() {
      const [code, response] = await GetRequest("upstreams");
      if (code !== 200) {
        toast.warning(`Unable to fetch upstreams`);
        return;
      }

      if (Array.isArray(response.upstreams)) {
        setUpstreams(response);
      } else {
        console.warn("Unexpected response format:", response);
      }
    }

    fetchupstreams();
  }, []);

  return (
    <div>
      <div className="flex gap-5">
        <AddUpstream />
      </div>
      {(upstreams && (
        <div className="grid lg:grid-cols-4 gap-2">
          {upstreams?.upstreams.map((upstream, index) => (
            <UpstreamCard
              key={index}
              dnsPing={upstream.dnsPing}
              icmpPing={upstream.icmpPing}
              name={upstream.name}
              preferred={upstream.preferred}
              upstream={upstream.upstream}
            />
          ))}
        </div>
      )) || (
        <div className="flex justify-center items-center">
          <div className="flex flex-col items-center space-y-4">
            <ArrowClockwise className="w-12 h-12 text-blue-400 animate-spin" />
            <p className="text-lg text-stone-400">Loading upstreams...</p>
          </div>
        </div>
      )}
    </div>
  );
}
