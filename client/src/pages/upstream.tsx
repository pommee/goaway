import { AddUpstream } from "@/app/upstream/addUpstream";
import { UpstreamCard } from "@/app/upstream/card";
import { GetRequest } from "@/util";
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
    </div>
  );
}
