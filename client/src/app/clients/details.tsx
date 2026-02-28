"use client";

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerClose
} from "@/components/ui/drawer";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { ClientEntry } from "@/pages/clients";
import { GetRequest, PutRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  ArrowsDownUpIcon,
  CaretDownIcon,
  CheckIcon,
  ClockCounterClockwiseIcon,
  LightningIcon,
  PencilIcon,
  PlusMinusIcon,
  RowsIcon,
  ShieldIcon,
  ShieldWarningIcon,
  SparkleIcon,
  TargetIcon,
  XIcon
} from "@phosphor-icons/react";
import { ReactNode, useEffect, useState } from "react";
import TimeAgo from "react-timeago";
import { toast } from "sonner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";

type AllDomains = { [domain: string]: number };

type ClientHistory = {
  domain: string;
  timestamp: Date;
};

type ClientEntryDetails = {
  allDomains: AllDomains;
  avgResponseTimeMs: number;
  blockedRequests: number;
  cachedRequests: number;
  mostQueriedDomain: string;
  totalRequests: number;
  uniqueDomains: number;
  clientInfo: {
    name: string;
    ip: string;
    mac: string;
    vendor: string;
    lastSeen: string;
    bypass: boolean;
  };
};

type CardDetailsProps = ClientEntry & {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function ClientDetails({
  open,
  onOpenChange,
  ...clientEntry
}: CardDetailsProps) {
  const [clientDetails, setClientDetails] = useState<ClientEntryDetails | null>(
    null
  );
  const [clientHistory, setClientHistory] = useState<ClientHistory[] | null>(
    null
  );
  const [isLoading, setIsLoading] = useState(true);

  const [isEditingName, setIsEditingName] = useState(false);
  const [editedName, setEditedName] = useState(clientEntry.name || "");
  const [updatingName, setUpdatingName] = useState(false);

  useEffect(() => {
    if (!open) return;

    async function fetchData() {
      setIsLoading(true);
      try {
        const [code, details] = await GetRequest(
          `client/${clientEntry.ip}/details`
        );
        if (code === 200) setClientDetails(details);

        const [hCode, hist] = await GetRequest(
          `client/${clientEntry.ip}/history`
        );
        if (hCode === 200) setClientHistory(hist.history);
      } catch {
        toast.error("Failed to load client data");
      } finally {
        setIsLoading(false);
      }
    }

    fetchData();
  }, [open, clientEntry.ip]);

  const updateClientBypass = async (enabled: boolean) => {
    try {
      const [code] = await PutRequest(
        `client/${clientEntry.ip}/bypass/${enabled}`,
        null,
        false
      );
      if (code !== 200) throw new Error();

      setClientDetails((prev) =>
        prev
          ? { ...prev, clientInfo: { ...prev.clientInfo, bypass: enabled } }
          : prev
      );
      toast.success(enabled ? "Bypass enabled" : "Bypass disabled");
    } catch {
      toast.error("Could not update bypass setting");
    }
  };

  const updateClientName = async () => {
    const newName = editedName.trim();
    if (!newName) {
      toast.warning("Name cannot be empty");
      setIsEditingName(false);
      return;
    }
    if (newName === (clientDetails?.clientInfo.name ?? clientEntry.name)) {
      setIsEditingName(false);
      return;
    }

    setUpdatingName(true);
    try {
      const [code] = await PutRequest(
        `client/${clientEntry.ip}/name/${encodeURIComponent(newName)}`,
        null,
        false
      );
      if (code === 200) {
        setClientDetails((prev) =>
          prev
            ? { ...prev, clientInfo: { ...prev.clientInfo, name: newName } }
            : prev
        );
        toast.success(`Name updated to "${newName}"`);
        setIsEditingName(false);
      } else {
        toast.error("Failed to update name");
      }
    } catch {
      toast.error("Error updating name");
    } finally {
      setUpdatingName(false);
    }
  };

  const getProgressColor = (value: number, max: number) => {
    const pct = (value / max) * 100;
    if (pct < 30) return "bg-muted-foreground";
    if (pct < 70) return "bg-muted-foreground";
    return "bg-muted-foreground";
  };

  const displayName =
    clientDetails?.clientInfo.name ?? clientEntry.name ?? "Unknown";

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent>
        <DrawerHeader>
          <div className="flex">
            <div className="flex-1">
              {isEditingName ? (
                <div className="flex">
                  <Input
                    value={editedName}
                    onChange={(e) => setEditedName(e.target.value)}
                    autoFocus
                    disabled={updatingName}
                    className="h-9"
                  />
                  <Button
                    size="icon"
                    variant="ghost"
                    onClick={updateClientName}
                    disabled={updatingName}
                  >
                    {updatingName ? (
                      <ArrowsClockwiseIcon className="h-4 w-4 animate-spin" />
                    ) : (
                      <CheckIcon className="h-4 w-4 text-green-500" />
                    )}
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    onClick={() => {
                      setEditedName(displayName);
                      setIsEditingName(false);
                    }}
                  >
                    <XIcon className="h-4 w-4 text-red-500" />
                  </Button>
                </div>
              ) : (
                <div className="group flex items-center gap-2">
                  <DrawerTitle className="text-xl font-semibold truncate leading-tight">
                    {displayName}
                  </DrawerTitle>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 opacity-0 group-hover:opacity-100"
                    onClick={() => {
                      setEditedName(displayName);
                      setIsEditingName(true);
                    }}
                  >
                    <PencilIcon size={14} className="text-muted-foreground" />
                  </Button>
                </div>
              )}

              <div className="mt-1.5 flex flex-wrap gap-2 text-xs">
                <div className="rounded bg-muted px-2 py-0.5 font-mono text-muted-foreground">
                  {clientEntry.ip}
                </div>
                {clientEntry.mac && (
                  <div className="rounded bg-muted px-2 py-0.5 font-mono text-muted-foreground">
                    {clientEntry.mac}
                  </div>
                )}
                {clientEntry.vendor && (
                  <div className="rounded bg-muted px-2 py-0.5 font-mono text-muted-foreground">
                    {clientEntry.vendor}
                  </div>
                )}
              </div>
            </div>

            <DrawerClose asChild>
              <Button variant="ghost" size="icon" className="shrink-0">
                <XIcon size={20} />
              </Button>
            </DrawerClose>
          </div>

          {clientEntry.lastSeen && (
            <div className="mt-2 text-xs text-muted-foreground">
              Last seen <TimeAgo date={clientEntry.lastSeen} maxPeriod={60} />
            </div>
          )}
        </DrawerHeader>

        <div className="flex flex-col flex-1">
          <Tabs defaultValue="overview" className="flex flex-col flex-1">
            <div className="border-b">
              <TabsList className="w-full justify-start gap-5 bg-transparent p-0">
                <TabsTrigger
                  value="overview"
                  className="gap-0 rounded-none data-[state=active]:bg-transparent! border-none hover:cursor-pointer hover:font-semibold"
                >
                  <TargetIcon size={16} className="mr-2" />
                  Overview
                </TabsTrigger>
                <TabsTrigger
                  value="domains"
                  className="gap-0 rounded-none data-[state=active]:bg-transparent! border-none hover:cursor-pointer hover:font-semibold"
                >
                  <RowsIcon size={16} className="mr-2" />
                  Domains
                </TabsTrigger>
                <TabsTrigger
                  value="history"
                  className="gap-0 rounded-none data-[state=active]:bg-transparent! border-none hover:cursor-pointer hover:font-semibold"
                >
                  <ClockCounterClockwiseIcon size={16} className="mr-2" />
                  History
                </TabsTrigger>
              </TabsList>
            </div>

            {isLoading ? (
              <div className="flex flex-1 items-center justify-center">
                <div className="h-10 w-10 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              </div>
            ) : clientDetails ? (
              <>
                <TabsContent
                  value="overview"
                  className="mt-0 flex-1 overflow-y-auto px-2 py-2"
                >
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-2">
                    <StatTile
                      icon={<ArrowsDownUpIcon />}
                      label="Total Requests"
                      value={clientDetails.totalRequests.toLocaleString()}
                    />
                    <StatTile
                      icon={<SparkleIcon />}
                      label="Unique Domains"
                      value={clientDetails.uniqueDomains.toLocaleString()}
                    />
                    <StatTile
                      icon={<ShieldWarningIcon />}
                      label="Blocked"
                      value={clientDetails.blockedRequests.toLocaleString()}
                      accent="text-red-500"
                    />
                    <StatTile
                      icon={<LightningIcon />}
                      label="Cached"
                      value={clientDetails.cachedRequests.toLocaleString()}
                      accent="text-green-500"
                    />
                    <StatTile
                      icon={<PlusMinusIcon />}
                      label="Avg Response"
                      value={`${clientDetails.avgResponseTimeMs.toLocaleString()} ms`}
                    />
                    <StatTile
                      icon={<CaretDownIcon />}
                      label="Top Domain"
                      value={clientDetails.mostQueriedDomain}
                      truncate
                    />
                  </div>

                  <Separator className="my-4" />

                  <h3 className="mb-2 text-sm font-medium">
                    Top Queried Domains
                  </h3>
                  <div className="space-y-2">
                    {Object.entries(clientDetails.allDomains)
                      .sort(([, a], [, b]) => b - a)
                      .slice(0, 6)
                      .map(([domain, count]) => {
                        const max = Math.max(
                          ...Object.values(clientDetails.allDomains)
                        );
                        return (
                          <div key={domain} className="space-y-1">
                            <div className="flex items-center justify-between text-sm">
                              <span className="truncate font-medium">
                                {domain}
                              </span>
                              <span className="tabular-nums text-muted-foreground">
                                {count}
                              </span>
                            </div>
                            <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                              <div
                                className={cn(
                                  "h-full rounded",
                                  getProgressColor(count, max)
                                )}
                                style={{ width: `${(count / max) * 100}%` }}
                              />
                            </div>
                          </div>
                        );
                      })}
                  </div>

                  <Separator className="my-4" />

                  <div className="rounded-lg border bg-muted/40 p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">Bypass Filtering</div>
                        <div className="text-xs text-muted-foreground">
                          Allow this client to bypass blocklists
                        </div>
                      </div>
                      <Switch
                        checked={clientDetails.clientInfo.bypass}
                        onCheckedChange={updateClientBypass}
                      />
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="domains" className="mt-0 flex-1">
                  <div className="px-1 mb-2 flex items-center justify-between">
                    <h3 className="text-sm font-medium">All Queried Domains</h3>
                    <span className="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium">
                      {Object.keys(clientDetails.allDomains).length} domains
                    </span>
                  </div>

                  <ScrollArea type="always" className="h-[calc(100vh-200px)]">
                    <Table>
                      <TableHeader className="sticky top-0 z-10 bg-muted/80">
                        <TableRow>
                          <TableHead className="w-20 text-left">
                            Count
                          </TableHead>
                          <TableHead>Domain</TableHead>
                          <TableHead className="w-24 text-right">
                            % of total
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {Object.entries(clientDetails.allDomains)
                          .sort(([, a], [, b]) => b - a)
                          .map(([domain, count]) => {
                            const percentage = (
                              (count / clientDetails.totalRequests) *
                              100
                            ).toFixed(1);
                            return (
                              <TableRow
                                key={domain}
                                className="hover:bg-muted/50 transition-colors"
                              >
                                <TableCell className="text-left font-medium tabular-nums">
                                  {count}
                                </TableCell>
                                <TableCell className="max-w-24 truncate">
                                  {domain}
                                </TableCell>
                                <TableCell className="text-right text-muted-foreground tabular-nums">
                                  {percentage}%
                                </TableCell>
                              </TableRow>
                            );
                          })}
                      </TableBody>
                    </Table>
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="history" className="flex-1 overflow-auto">
                  {clientHistory?.length ? (
                    <ScrollArea type="always" className="h-[calc(100vh-170px)]">
                      <Table className="border-separate border-spacing-0">
                        <TableHeader className="bg-muted/80">
                          <TableRow>
                            <TableHead>Domain</TableHead>
                            <TableHead className="text-right">When</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {clientHistory.map(
                            (entry: ClientHistory, i: number) => (
                              <TableRow
                                key={i}
                                className={cn(
                                  "transition-colors",
                                  i % 2 === 0 ? "bg-muted/20" : "bg-background"
                                )}
                              >
                                <TableCell className="max-w-48 truncate">
                                  {entry.domain}
                                </TableCell>
                                <TableCell className="text-right text-xs text-muted-foreground whitespace-nowrap">
                                  <TimeAgo
                                    date={new Date(entry.timestamp)}
                                    minPeriod={60}
                                  />
                                </TableCell>
                              </TableRow>
                            )
                          )}
                        </TableBody>
                      </Table>
                    </ScrollArea>
                  ) : (
                    <div className="flex h-full flex-col items-center justify-center text-center text-muted-foreground">
                      <ClockCounterClockwiseIcon
                        size={40}
                        className="mb-4 opacity-50"
                      />
                      <p className="text-sm">No recent DNS history available</p>
                    </div>
                  )}
                </TabsContent>
              </>
            ) : (
              <div className="flex flex-1 flex-col items-center justify-center text-muted-foreground">
                <ShieldIcon size={48} className="mb-4 opacity-60" />
                <p className="text-lg font-medium">No client data available</p>
              </div>
            )}
          </Tabs>
        </div>
      </DrawerContent>
    </Drawer>
  );
}

function StatTile({
  icon,
  label,
  value,
  accent = "text-foreground",
  truncate = false
}: {
  icon: ReactNode;
  label: string;
  value: string;
  accent?: string;
  truncate?: boolean;
}) {
  return (
    <div className="rounded-lg border bg-card p-2 shadow-sm">
      <div className="mb-1 flex items-center gap-1.5 text-xs text-muted-foreground">
        {icon}
        {label}
      </div>
      <div className={cn("font-semibold", accent, truncate && "truncate")}>
        {value}
      </div>
    </div>
  );
}
