import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { ClientEntry } from "@/pages/clients";
import { GetRequest } from "@/util";
import { useState, useEffect } from "react";
import { toast } from "sonner";

type AllDomains = {
  name: string;
  amount: number;
};

type ClientEntryDetails = {
  allDomains: AllDomains;
  avgResponseTimeMs: number;
  blockedRequests: number;
  cachedRequests: number;
  ip: string;
  lastSeen: string;
  mostQueriedDomain: string;
  totalRequests: number;
  uniqueDomains: number;
};

export function CardDetails({
  onClose,
  ...clientEntry
}: ClientEntry & { onClose: () => void }) {
  const [clientDetails, setClientDetails] = useState<ClientEntryDetails | null>(
    null
  );

  useEffect(() => {
    async function fetchClients() {
      const [code, response] = await GetRequest(
        `clientDetails?clientIP=${clientEntry.ip}`
      );
      if (code !== 200) {
        toast.warning(`Unable to fetch client details`);
        return;
      }

      setClientDetails(response.details);
    }

    fetchClients();
  }, [clientEntry.ip]);

  if (!clientDetails) {
    return <div></div>;
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="bg-zinc-900 text-white border-zinc-800 rounded-xl w-1/2 max-w-none">
        <DialogTitle className="text-center">
          <h2 className="text-2xl mb-1">{clientEntry.name}</h2>
        </DialogTitle>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm text-gray-400">
          <div>IP: {clientEntry.ip}</div>
          <div>MAC: {clientEntry.mac || "unknown"}</div>
          <div>Vendor: {clientEntry.vendor || "unknown"}</div>
          <div>Last Seen: {clientEntry.lastSeen}</div>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Total Requests</div>
            <div className="text-2xl">{clientDetails.totalRequests}</div>
          </div>
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Unique Domains</div>
            <div className="text-2xl">{clientDetails.uniqueDomains}</div>
          </div>
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Blocked Requests</div>
            <div className="text-2xl">{clientDetails.blockedRequests}</div>
          </div>
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Cached Requests</div>
            <div className="text-2xl">{clientDetails.cachedRequests}</div>
          </div>
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Avg Response Time</div>
            <div className="text-2xl">{clientDetails.avgResponseTimeMs} ms</div>
          </div>
          <div className="bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-gray-400">Most Queried</div>
            <div className="text-2xl truncate">
              {clientDetails.mostQueriedDomain}
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-center text-lg font-semibold mb-3">
            All Queried Domains
          </h3>
          <div className="bg-gray-800 rounded-lg p-2 max-h-64 overflow-y-auto">
            {Object.entries(clientDetails.allDomains)
              .sort((a, b) => b[1] - a[1])
              .map(([domain, count], index) => (
                <div
                  key={index}
                  className="flex justify-between py-1 px-2 hover:bg-gray-700 rounded"
                >
                  <div className="text-gray-300 w-8 text-left mr-3">
                    {count}
                  </div>
                  <div className="flex-1 truncate">{domain}</div>
                </div>
              ))}
          </div>
        </div>

        <div className="flex justify-between mt-6">
          <button className="bg-blue-500 hover:bg-blue-600 px-4 py-2 rounded text-sm">
            Details
          </button>
          <button className="bg-blue-500 hover:bg-blue-600 px-4 py-2 rounded text-sm">
            Block Device
          </button>
          <button className="bg-blue-500 hover:bg-blue-600 px-4 py-2 rounded text-sm">
            Settings
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
