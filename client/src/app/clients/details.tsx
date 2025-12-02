import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { ClientEntry } from "@/pages/clients";
import { GetRequest } from "@/util";
import {
  CaretDownIcon,
  ClockCounterClockwiseIcon,
  EyeglassesIcon,
  LightningIcon,
  PlusMinusIcon,
  RowsIcon,
  ShieldIcon,
  SparkleIcon,
  TargetIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import TimeAgo from "react-timeago";
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
  const [clientHistory, setClientHistory] = useState(null);

  useEffect(() => {
    async function fetchClient() {
      setIsLoading(true);
      try {
        const [code, response] = await GetRequest(
          `clientDetails?clientIP=${clientEntry.ip}`
        );
        if (code !== 200) {
          toast.warning("Unable to fetch client details");
          return;
        }

        setClientDetails(response);
      } catch {
        toast.error("Error fetching client details");
      } finally {
        setIsLoading(false);
      }
    }

    fetchClient();
  }, [clientEntry.ip]);

  const getProgressColor = (value: number, max: number) => {
    const percentage = (value / max) * 100;
    if (percentage < 30) return "bg-emerald-500";
    if (percentage < 70) return "bg-amber-500";
    return "bg-red-500";
  };

  useEffect(() => {
    async function getClientHistory() {
      try {
        const [code, response] = await GetRequest(
          `clientHistory?ip=${clientEntry.ip}`
        );
        if (code !== 200) {
          toast.warning("Unable to fetch client history");
          return;
        }

        setClientHistory(response.history);
      } catch {
        toast.error("Error fetching client history");
      } finally {
        setIsLoading(false);
      }
    }

    getClientHistory();
  }, [clientEntry.ip]);

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="border-none bg-accent rounded-lg w-full max-w-6xl max-h-3/4 overflow-y-auto">
        <DialogTitle>
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl sm:text-2xl font-bold mb-1">
                {clientEntry.name || "unknown"}
              </h2>
              <div className="flex items-center text-sm gap-2">
                <span className="bg-muted-foreground/20 px-2 py-0.5 rounded-md font-mono text-xs">
                  ip: {clientEntry.ip}
                </span>
                {clientEntry.mac && (
                  <span className="bg-muted-foreground/20 px-2 py-0.5 rounded-md font-mono text-xs">
                    mac: {clientEntry.mac}
                  </span>
                )}
                {clientEntry.vendor && (
                  <span className="bg-muted-foreground/20 px-2 py-0.5 rounded-md font-mono text-xs">
                    vendor: {clientEntry.vendor}
                  </span>
                )}
              </div>
            </div>
            <div className="text-right hidden sm:block">
              <span className="text-xs">Last Activity</span>
              <div className="text-muted-foreground">
                {clientEntry.lastSeen}
              </div>
            </div>
          </div>
        </DialogTitle>

        <Tabs defaultValue="overview">
          <TabsList className="bg-transparent space-x-2">
            <TabsTrigger
              value="overview"
              className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
            >
              <TargetIcon />
              Overview
            </TabsTrigger>
            <TabsTrigger
              value="domains"
              className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
            >
              <RowsIcon />
              Domains
            </TabsTrigger>
            <TabsTrigger
              value="history"
              className="border-l-0 !bg-transparent border-t-0 border-r-0 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:!border-b-primary rounded-none"
            >
              <ClockCounterClockwiseIcon />
              History
            </TabsTrigger>
          </TabsList>

          {isLoading ? (
            <div className="flex justify-center items-center p-16">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500" />
            </div>
          ) : clientDetails ? (
            <div>
              <TabsContent value="overview">
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2 mb-4">
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
                    value={clientDetails.mostQueriedDomain}
                    color="bg-indigo-600"
                  />
                </div>

                <div className="mb-6">
                  <p className="mb-2">Top Queried Domains</p>
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
                            className="bg-background rounded-md overflow-hidden shadow-sm"
                          >
                            <div className="flex items-center p-2">
                              <div className="w-12 text-center font-mono py-1 rounded text-xs font-medium">
                                {count}
                              </div>
                              <div className="ml-3 grow font-medium truncate">
                                {domain}
                              </div>
                              <div className="w-24 shrink-0">
                                <div className="h-2 bg-accent rounded-full w-full">
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
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-4">
                  <Button variant="outline" disabled={true}>
                    [WIP] View Details
                  </Button>
                  <Button variant="outline" disabled={true}>
                    [WIP] Block Device
                  </Button>
                  <Button variant="outline" disabled={true}>
                    [WIP] Device Settings
                  </Button>
                </div>
              </TabsContent>
              <TabsContent value="domains">
                <p className="mb-2">
                  All Queried Domains
                  <span className="ml-2 text-xs bg-accent px-2 py-0.5 rounded-full text-muted-foreground">
                    {Object.keys(clientDetails.allDomains).length} domains
                  </span>
                </p>

                <div className="shadow-md border rounded-md overflow-hidden">
                  <div className="flex justify-between items-center py-2 px-3">
                    <div className="w-16 text-xs text-muted-foreground font-medium">
                      Count
                    </div>
                    <div className="grow text-xs text-muted-foreground font-medium">
                      Domain
                    </div>
                    <div className="w-24 text-xs text-muted-foreground font-medium">
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
                            className="flex items-center py-2 px-3 hover:bg-accent border-b border-accent last:border-0"
                          >
                            <div className="w-16 font-mono bg-accent py-1 rounded text-center text-xs font-medium">
                              {count}
                            </div>
                            <div className="ml-3 grow font-medium truncate">
                              {domain}
                            </div>
                            <div className="w-24 text-right text-muted-foreground text-sm">
                              {percentage}%
                            </div>
                          </div>
                        );
                      })}
                  </div>
                </div>
              </TabsContent>
              <TabsContent value="history">
                {Array.isArray(clientHistory) && clientHistory.length > 0 ? (
                  <div>
                    {clientHistory.map((entry, index) => (
                      <div
                        key={index}
                        className={cn(
                          "flex justify-between px-3 py-2 hover:font-bold",
                          index % 2 === 0
                            ? "bg-secondary"
                            : "bg-muted-foreground/10"
                        )}
                      >
                        <span className="truncate">{entry.domain}</span>
                        <span className="text-xs text-muted-foreground">
                          <TimeAgo
                            date={new Date(entry.timestamp)}
                            minPeriod={60}
                          />
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground text-sm">
                    No history available.
                  </p>
                )}
              </TabsContent>
            </div>
          ) : (
            <div className="text-center py-16 grow flex flex-col items-center justify-center">
              <ShieldIcon size={48} className="mb-4" />
              <div className="text-lg">No data available for this client</div>
              <div className="text-sm mt-2 text-muted-foreground">
                Try checking the network connection or refreshing
              </div>
            </div>
          )}
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

function StatCard({ icon, label, value, color }) {
  return (
    <div className="rounded-sm shadow-md bg-background">
      <div className={`${color} h-1`}></div>
      <div className="p-2">
        <div className="flex items-center text-xs text-muted-foreground mb-1 gap-1">
          {icon} {label}
        </div>
        <div className="font-bold text-sm truncate">{value}</div>
      </div>
    </div>
  );
}
