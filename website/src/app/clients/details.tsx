import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { ClientEntry } from "@/pages/clients";
import { GetRequest } from "@/util";
import { useState, useEffect } from "react";
import { toast } from "sonner";

type AllDomains = {
  [domain: string]: number;
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
    null,
  );
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchClients() {
      setIsLoading(true);
      try {
        const [code, response] = await GetRequest(
          `clientDetails?clientIP=${clientEntry.ip}`,
        );
        if (code !== 200) {
          toast.warning(`Unable to fetch client details`);
          return;
        }

        setClientDetails(response.details);
      } catch {
        toast.error("Error fetching client details");
      } finally {
        setIsLoading(false);
      }
    }

    fetchClients();
  }, [clientEntry.ip]);

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="bg-zinc-900 text-white border-zinc-800 rounded-xl w-full max-w-4xl mx-auto p-4 sm:p-6 overflow-hidden max-h-[90vh] flex flex-col">
        <DialogTitle className="text-center shrink-0">
          <h2 className="text-xl sm:text-2xl font-bold text-blue-400 mb-2">
            {clientEntry.name}
          </h2>
        </DialogTitle>

        {isLoading ? (
          <div className="flex justify-center items-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        ) : clientDetails ? (
          <div className="overflow-y-auto pr-1 -mr-1 pb-4 flex-grow">
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3 text-sm text-gray-300 bg-zinc-800 p-3 rounded-lg mb-4">
              <div className="flex items-center space-x-2">
                <span className="text-gray-400">IP:</span>
                <span>{clientEntry.ip}</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="text-gray-400">MAC:</span>
                <span>{clientEntry.mac || "unknown"}</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="text-gray-400">Vendor:</span>
                <span>{clientEntry.vendor || "unknown"}</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="text-gray-400">Last Seen:</span>
                <span>{clientEntry.lastSeen}</span>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mb-6">
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">Total Requests</div>
                <div className="text-xl font-bold text-blue-400">
                  {clientDetails.totalRequests.toLocaleString()}
                </div>
              </div>
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">Unique Domains</div>
                <div className="text-xl font-bold text-blue-400">
                  {clientDetails.uniqueDomains.toLocaleString()}
                </div>
              </div>
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">
                  Blocked Requests
                </div>
                <div className="text-xl font-bold text-blue-400">
                  {clientDetails.blockedRequests.toLocaleString()}
                </div>
              </div>
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">
                  Cached Requests
                </div>
                <div className="text-xl font-bold text-blue-400">
                  {clientDetails.cachedRequests.toLocaleString()}
                </div>
              </div>
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">
                  Avg Response Time
                </div>
                <div className="text-xl font-bold text-blue-400">
                  {clientDetails.avgResponseTimeMs.toLocaleString()} ms
                </div>
              </div>
              <div className="bg-zinc-800 p-3 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md">
                <div className="text-sm text-gray-400 mb-1">Most Queried</div>
                <div
                  className="text-xl font-bold text-blue-400 truncate"
                  title={clientDetails.mostQueriedDomain}
                >
                  {clientDetails.mostQueriedDomain}
                </div>
              </div>
            </div>

            <div className="mb-4">
              <h3 className="text-center text-lg font-semibold mb-3 text-blue-400">
                All Queried Domains
              </h3>
              <div className="bg-zinc-800 rounded-lg p-2 max-h-48 sm:max-h-56 overflow-y-auto border border-zinc-700 shadow-inner">
                {Object.entries(clientDetails.allDomains)
                  .sort((a, b) => b[1] - a[1])
                  .map(([domain, count], index) => (
                    <div
                      key={index}
                      className="flex justify-between py-2 px-2 hover:bg-zinc-700 rounded transition-colors"
                    >
                      <div className="bg-blue-500 text-white rounded px-2 py-1 text-xs font-medium w-10 text-center">
                        {count}
                      </div>
                      <div className="flex-1 truncate ml-2 text-gray-200">
                        {domain}
                      </div>
                    </div>
                  ))}
              </div>
            </div>

            <div className="flex flex-col sm:flex-row justify-between gap-2 mt-4">
              <button className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex-1">
                View Details
              </button>
              <button className="bg-red-600 hover:bg-red-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex-1">
                Block Device
              </button>
              <button className="bg-gray-600 hover:bg-gray-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex-1">
                Device Settings
              </button>
            </div>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-400">
            No data available for this client
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
