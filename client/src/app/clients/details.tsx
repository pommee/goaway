import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { ClientEntry } from "@/pages/clients";
import { GetRequest } from "@/util";
import {
  BirdIcon,
  CaretDownIcon,
  CaretRightIcon,
  EyeIcon,
  EyeglassesIcon,
  IdentificationBadgeIcon,
  LightningIcon,
  LineSegmentsIcon,
  PlusMinusIcon,
  ShieldIcon,
  SparkleIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
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
    null
  );
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("overview");

  useEffect(() => {
    async function fetchClients() {
      setIsLoading(true);
      try {
        const [code, response] = await GetRequest(
          `clientDetails?clientIP=${clientEntry.ip}`
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

  const getProgressColor = (value: number, max: number) => {
    const percentage = (value / max) * 100;
    if (percentage < 30) return "bg-emerald-500";
    if (percentage < 70) return "bg-amber-500";
    return "bg-red-500";
  };

  const formatTimeAgo = (dateString: string) => {
    return dateString;
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="bg-zinc-900 text-white border-zinc-700 rounded-lg w-full max-w-5xl mx-auto p-0 overflow-hidden max-h-[90vh] flex flex-col">
        <div className="bg-gradient-to-r from-blue-900 to-indigo-900 p-4 sm:p-6 border-b border-zinc-700">
          <DialogTitle>
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-2xl sm:text-3xl font-bold text-white mb-1">
                  {clientEntry.name || "Unnamed Device"}
                </h2>
                <div className="flex items-center text-blue-200 text-sm">
                  <span className="bg-blue-800 px-2 py-0.5 rounded-full font-medium">
                    {clientEntry.ip}
                  </span>
                  {clientEntry.mac && (
                    <span className="ml-2 flex items-center">
                      <IdentificationBadgeIcon size={14} className="mr-1" />
                      {clientEntry.mac}
                    </span>
                  )}
                  {clientEntry.vendor && (
                    <span className="ml-2 opacity-75">
                      {clientEntry.vendor}
                    </span>
                  )}
                </div>
              </div>
              <div className="text-right hidden sm:block">
                <span className="text-xs text-blue-200">Last Activity</span>
                <div className="text-lg font-medium">
                  {formatTimeAgo(clientEntry.lastSeen)}
                </div>
              </div>
            </div>
          </DialogTitle>
        </div>

        <div className="flex bg-zinc-800 border-b border-zinc-700">
          <button
            className={`px-4 py-2 text-sm font-medium flex items-center ${
              activeTab === "overview"
                ? "bg-zinc-900 text-blue-400 border-b-2 border-blue-500"
                : "text-zinc-400 hover:bg-zinc-700 hover:text-white"
            }`}
            onClick={() => setActiveTab("overview")}
          >
            <BirdIcon size={16} className="mr-2" />
            Overview
          </button>
          <button
            className={`px-4 py-2 text-sm font-medium flex items-center ${
              activeTab === "domains"
                ? "bg-zinc-900 text-blue-400 border-b-2 border-blue-500"
                : "text-zinc-400 hover:bg-zinc-700 hover:text-white"
            }`}
            onClick={() => setActiveTab("domains")}
          >
            <LineSegmentsIcon size={16} className="mr-2" />
            Domains
          </button>
        </div>

        {isLoading ? (
          <div className="flex justify-center items-center p-16">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        ) : clientDetails ? (
          <div className="overflow-y-auto p-4 sm:p-6 flex-grow">
            {activeTab === "overview" && (
              <>
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
                  <StatCard
                    icon={<EyeglassesIcon size={18} />}
                    label="Total Requests"
                    value={clientDetails.totalRequests.toLocaleString()}
                    color="bg-blue-600"
                  />
                  <StatCard
                    icon={<SparkleIcon size={18} />}
                    label="Unique Domains"
                    value={clientDetails.uniqueDomains.toLocaleString()}
                    color="bg-purple-600"
                  />
                  <StatCard
                    icon={<ShieldIcon size={18} />}
                    label="Blocked Requests"
                    value={clientDetails.blockedRequests.toLocaleString()}
                    color="bg-red-600"
                  />
                  <StatCard
                    icon={<LightningIcon size={18} />}
                    label="Cached Requests"
                    value={clientDetails.cachedRequests.toLocaleString()}
                    color="bg-emerald-600"
                  />
                  <StatCard
                    icon={<PlusMinusIcon size={18} />}
                    label="Avg Response"
                    value={`${clientDetails.avgResponseTimeMs.toLocaleString()} ms`}
                    color="bg-amber-600"
                  />
                  <StatCard
                    icon={<CaretDownIcon size={18} />}
                    label="Most Queried"
                    value={clientDetails.mostQueriedDomain.split(".")[0]}
                    color="bg-indigo-600"
                  />
                </div>

                <div className="mb-6">
                  <h3 className="text-lg font-semibold mb-3 text-white flex items-center">
                    <EyeIcon size={18} className="mr-2 text-blue-400" />
                    Top Queried Domains
                  </h3>
                  <div className="grid gap-2">
                    {Object.entries(clientDetails.allDomains)
                      .sort((a, b) => b[1] - a[1])
                      .slice(0, 5)
                      .map(([domain, count], index) => {
                        const max = Math.max(
                          ...Object.values(clientDetails.allDomains)
                        );
                        return (
                          <div
                            key={index}
                            className="bg-zinc-800 rounded-md overflow-hidden border border-zinc-700 hover:border-blue-500 transition-colors"
                          >
                            <div className="flex items-center p-2">
                              <div className="w-12 text-center font-mono bg-zinc-900 py-1 rounded text-xs font-medium">
                                {count}
                              </div>
                              <div className="ml-3 flex-grow font-medium truncate text-zinc-200">
                                {domain}
                              </div>
                              <div className="w-24 flex-shrink-0">
                                <div className="h-2 bg-zinc-700 rounded-full w-full">
                                  <div
                                    className={`h-2 rounded-full ${getProgressColor(
                                      count,
                                      max
                                    )}`}
                                    style={{ width: `${(count / max) * 100}%` }}
                                  ></div>
                                </div>
                              </div>
                            </div>
                          </div>
                        );
                      })}
                  </div>
                  <div className="mt-2 text-center">
                    <button
                      className="text-blue-400 hover:text-blue-300 text-sm flex items-center mx-auto"
                      onClick={() => setActiveTab("domains")}
                    >
                      View all domains <CaretRightIcon size={16} />
                    </button>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-4">
                  <ActionButton
                    label="View Details"
                    bgClass="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800"
                  />
                  <ActionButton
                    label="Block Device"
                    bgClass="bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800"
                  />
                  <ActionButton
                    label="Device Settings"
                    bgClass="bg-gradient-to-r from-zinc-600 to-zinc-700 hover:from-zinc-700 hover:to-zinc-800"
                  />
                </div>
              </>
            )}

            {activeTab === "domains" && (
              <div>
                <h3 className="text-lg font-semibold mb-3 text-white flex items-center">
                  <Eye size={18} className="mr-2 text-blue-400" />
                  All Queried Domains
                  <span className="ml-2 text-xs bg-zinc-800 px-2 py-0.5 rounded-full text-zinc-400">
                    {Object.keys(clientDetails.allDomains).length} domains
                  </span>
                </h3>

                <div className="bg-zinc-800/50 border border-zinc-700 rounded-md overflow-hidden">
                  <div className="flex justify-between items-center py-2 px-3 bg-zinc-800 border-b border-zinc-700">
                    <div className="w-16 text-xs text-zinc-400 font-medium">
                      Count
                    </div>
                    <div className="flex-grow text-xs text-zinc-400 font-medium">
                      Domain
                    </div>
                    <div className="w-24 text-xs text-zinc-400 font-medium">
                      Percentage
                    </div>
                  </div>

                  <div className="max-h-96 overflow-y-auto">
                    {Object.entries(clientDetails.allDomains)
                      .sort((a, b) => b[1] - a[1])
                      .map(([domain, count], index) => {
                        const totalRequests = clientDetails.totalRequests;
                        const percentage = (
                          (count / totalRequests) *
                          100
                        ).toFixed(1);
                        return (
                          <div
                            key={index}
                            className="flex items-center py-2 px-3 hover:bg-zinc-700/50 border-b border-zinc-700/50 last:border-0"
                          >
                            <div className="w-16 font-mono bg-zinc-900 py-1 rounded text-center text-xs font-medium">
                              {count}
                            </div>
                            <div className="ml-3 flex-grow font-medium truncate text-zinc-200">
                              {domain}
                            </div>
                            <div className="w-24 text-right text-zinc-400 text-sm">
                              {percentage}%
                            </div>
                          </div>
                        );
                      })}
                  </div>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-16 text-gray-400 flex-grow flex flex-col items-center justify-center">
            <ShieldIcon size={48} className="mb-4 text-zinc-600" />
            <div className="text-lg">No data available for this client</div>
            <div className="text-sm mt-2 text-zinc-500">
              Try checking the network connection or refreshing
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

function StatCard({ icon, label, value, color }) {
  return (
    <div className="bg-zinc-800 rounded-lg border border-zinc-700 hover:border-blue-500 transition-colors shadow-md overflow-hidden">
      <div className={`${color} h-1`}></div>
      <div className="p-3">
        <div className="flex items-center text-xs text-zinc-400 mb-1">
          <span className="mr-1">{icon}</span>
          {label}
        </div>
        <div className="font-bold text-lg text-white truncate">{value}</div>
      </div>
    </div>
  );
}

function ActionButton({ label, bgClass }) {
  return (
    <button
      className={`${bgClass} px-4 py-3 rounded-md text-sm font-medium transition-all shadow-md hover:shadow-lg`}
    >
      {label}
    </button>
  );
}
