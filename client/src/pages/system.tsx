"use client";

import { useState, useEffect } from "react";
import { ResponsiveContainer, PieChart, Pie, Cell } from "recharts";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { GetRequest } from "@/util";
import {
  ArrowClockwise,
  Clock,
  ComputerTower,
  GitCommit,
  Memory,
  Network,
  Pipe,
  HardDrives,
  Cpu,
  Desktop,
  Globe,
  Info,
  ChartLine,
  Barcode,
  WifiHigh
} from "@phosphor-icons/react";
import { Progress } from "@/components/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";

export function System() {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    cpuUsage: 12,
    memoryUsed: 12,
    memoryTotal: 16,
    uptime: "4d 12h 47m",
    version: "v1.14.0",
    commit: "12hjhu2u213hu",
    lastUpdate: "2024-04-27 13:37",
    services: [
      { name: "DNS Server", port: "6121", status: "online" },
      { name: "Web Server", port: "8080", status: "online" },
      { name: "Database", port: "local", status: "healthy" }
    ]
  });

  const [apiData, setApiData] = useState();
  const [storageData, setStorageData] = useState([]);
  const COLORS = ["#3B82F6", "#10B981", "#F59E0B", "#6366F1"];

  useEffect(() => {
    fetchStatistics();
    const interval = setInterval(fetchStatistics, 1000);
    return () => clearInterval(interval);
  }, []);

  const fetchStatistics = async () => {
    setLoading(true);

    try {
      const [code, data] = await GetRequest("system");

      if (code === 200 && data) {
        const apiData: System = data.system;
        setApiData(apiData);

        setStorageData(
          apiData.storage.map((disk) => ({
            name: disk.name,
            value: disk.size,
            model: disk.model,
            serial: disk.serial
          }))
        );

        setStats((prev) => ({
          ...prev,
          systemInfo: apiData,
          version: apiData.sysinfo.version,
          lastUpdate: new Date(apiData.sysinfo.timestamp).toLocaleString()
        }));
      }

      setStats((prev) => ({
        ...prev
      }));
    } catch (error) {
      console.error("Error fetching system data:", error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case "online":
        return "text-green-500";
      case "healthy":
        return "text-green-500";
      default:
        return "text-red-500";
    }
  };

  return (
    <div className="space-y-6 py-4">
      <div className="flex justify-between items-center">
        <div className="flex flex-col">
          <h2 className="text-2xl font-bold text-gray-50">
            {stats.systemInfo?.node?.hostname || ""}
          </h2>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={fetchStatistics}
          disabled={loading}
          className="flex items-center gap-2"
        >
          <ArrowClockwise
            className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-gray-700 bg-gray-900">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-blue-400" />
              <CardTitle className="text-sm font-medium text-gray-200">
                CPU
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              <span className="text-2xl font-bold text-gray-50">
                {stats.cpuUsage}%
              </span>
              <span className="text-xs text-gray-400">
                {stats.systemInfo?.cpu?.model || "Loading..."}
              </span>
            </div>
          </CardContent>
          <CardFooter>
            <div className="text-xs text-gray-400">
              {stats.systemInfo?.cpu?.cores || 0} Cores /{" "}
              {stats.systemInfo?.cpu?.threads || 0} Threads
            </div>
          </CardFooter>
        </Card>

        <Card className="border-gray-700 bg-gray-900">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Memory className="h-4 w-4 text-purple-400" />
              <CardTitle className="text-sm font-medium text-gray-200">
                Memory
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              <span className="text-2xl font-bold text-gray-50">
                {stats.memoryUsed}GB
              </span>
              <span className="text-xs text-gray-400">
                of {stats.memoryTotal}GB (
                {Math.round((stats.memoryUsed / stats.memoryTotal) * 100)}%)
              </span>
            </div>
          </CardContent>
          <CardFooter>
            <Progress
              value={(stats.memoryUsed / stats.memoryTotal) * 100}
              className="h-2 bg-gray-700"
            />
          </CardFooter>
        </Card>

        <Card className="border-gray-700 bg-gray-900">
          <CardHeader>
            <div className="flex items-center gap-2">
              <HardDrives className="h-4 w-4 text-amber-400" />
              <CardTitle className="text-sm font-medium text-gray-200">
                Storage
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              <span className="text-2xl font-bold text-gray-50">
                {stats.systemInfo?.storage
                  ? `${stats.systemInfo.storage.reduce(
                      (acc, disk) => acc + disk.size,
                      0
                    )}GB`
                  : "Loading..."}
              </span>
              <span className="text-xs text-gray-400">
                {stats.systemInfo?.storage
                  ? `${stats.systemInfo.storage.length} Devices`
                  : ""}
              </span>
            </div>
          </CardContent>
          <CardFooter>
            <div className="text-xs text-gray-400 flex gap-2">
              {storageData.slice(0, 3).map((disk, index) => (
                <div key={index} className="flex items-center gap-1">
                  <div
                    className="w-2 h-2 rounded-full"
                    style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  ></div>
                  <span>{disk.name}</span>
                </div>
              ))}
            </div>
          </CardFooter>
        </Card>

        <Card className="border-gray-700 bg-gray-900">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Network className="h-4 w-4 text-green-400" />
              <CardTitle className="text-sm font-medium text-gray-200">
                Network
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              <span className="text-2xl font-bold text-gray-50">
                {stats.systemInfo?.network
                  ? `${stats.systemInfo.network.length} Interfaces`
                  : "Loading..."}
              </span>
              <span className="text-xs text-gray-400">
                {stats.systemInfo?.network?.[0]?.speed
                  ? `${stats.systemInfo.network[0].speed} Mbps`
                  : ""}
              </span>
            </div>
          </CardContent>
          <CardFooter>
            <div className="text-xs text-gray-400 flex gap-2">
              {stats.systemInfo?.network?.slice(0, 2).map((iface, index) => (
                <div key={index} className="flex items-center gap-1">
                  {iface.name.startsWith("w") ? (
                    <WifiHigh className="h-3 w-3" />
                  ) : (
                    <Globe className="h-3 w-3" />
                  )}
                  <span>{iface.name}</span>
                </div>
              ))}
            </div>
          </CardFooter>
        </Card>
      </div>

      <Tabs defaultValue="performance" className="w-full">
        <TabsList className="bg-gray-800 border border-gray-700">
          <TabsTrigger
            value="performance"
            className="data-[state=active]:bg-gray-700"
          >
            <ChartLine className="h-4 w-4 mr-2" />
            Performance
          </TabsTrigger>
          <TabsTrigger
            value="system"
            className="data-[state=active]:bg-gray-700"
          >
            <ComputerTower className="h-4 w-4 mr-2" />
            System Info
          </TabsTrigger>
          <TabsTrigger
            value="storage"
            className="data-[state=active]:bg-gray-700"
          >
            <HardDrives className="h-4 w-4 mr-2" />
            Storage
          </TabsTrigger>
          <TabsTrigger
            value="network"
            className="data-[state=active]:bg-gray-700"
          >
            <Network className="h-4 w-4 mr-2" />
            Network
          </TabsTrigger>
        </TabsList>

        <TabsContent value="performance" className="mt-4 space-y-4">
          <Card className="border-gray-700 bg-gray-900">
            <CardHeader className="pb-2">
              <div className="flex items-center gap-2">
                <Memory className="h-4 w-4 text-purple-400" />
                <CardTitle className="text-sm font-medium text-gray-200">
                  Memory Allocation
                </CardTitle>
              </div>
              <CardDescription className="text-gray-400">
                Real-time memory usage
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-16 w-full bg-gray-800" />
              ) : (
                <>
                  <div className="flex items-center space-x-2 mb-2">
                    <Progress
                      value={(stats.memoryUsed / stats.memoryTotal) * 100}
                      className="h-3 bg-gray-700"
                    />
                    <span className="text-xs text-gray-400">
                      {Math.round((stats.memoryUsed / stats.memoryTotal) * 100)}
                      %
                    </span>
                  </div>
                  <div className="grid grid-cols-3 gap-2 text-sm mt-4">
                    <div className="bg-gray-800/50 p-3 rounded-lg border border-gray-700">
                      <div className="text-xs text-gray-400">Used</div>
                      <div className="text-base font-medium text-gray-200">
                        {stats.memoryUsed}GB
                      </div>
                    </div>
                    <div className="bg-gray-800/50 p-3 rounded-lg border border-gray-700">
                      <div className="text-xs text-gray-400">Free</div>
                      <div className="text-base font-medium text-gray-200">
                        {stats.memoryTotal - stats.memoryUsed}GB
                      </div>
                    </div>
                    <div className="bg-gray-800/50 p-3 rounded-lg border border-gray-700">
                      <div className="text-xs text-gray-400">Total</div>
                      <div className="text-base font-medium text-gray-200">
                        {stats.memoryTotal}GB
                      </div>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="system" className="mt-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card className="border-gray-700 bg-gray-900">
              <CardHeader className="pb-2">
                <div className="flex items-center gap-2">
                  <ComputerTower className="h-4 w-4 text-amber-400" />
                  <CardTitle className="text-sm font-medium text-gray-200">
                    System Information
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <div className="space-y-3">
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                  </div>
                ) : (
                  <div className="space-y-3 text-sm">
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <ComputerTower className="h-4 w-4" /> Hostname
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.node?.hostname || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Desktop className="h-4 w-4" /> Operating System
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.os?.name || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Pipe className="h-4 w-4" /> Kernel
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.kernel?.release || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Clock className="h-4 w-4" /> Timezone
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.node?.timezone || "N/A"}
                      </span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="border-gray-700 bg-gray-900">
              <CardHeader className="pb-2">
                <div className="flex items-center gap-2">
                  <Cpu className="h-4 w-4 text-blue-400" />
                  <CardTitle className="text-sm font-medium text-gray-200">
                    Hardware Information
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <div className="space-y-3">
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                    <Skeleton className="h-5 w-full bg-gray-800" />
                  </div>
                ) : (
                  <div className="space-y-3 text-sm">
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Cpu className="h-4 w-4" /> CPU
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.cpu?.model || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Barcode className="h-4 w-4" /> Product
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.product?.name || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <Info className="h-4 w-4" /> Board
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.board?.name || "N/A"}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400 flex items-center gap-2">
                        <GitCommit className="h-4 w-4" /> BIOS
                      </span>
                      <span className="text-gray-300">
                        {stats.systemInfo?.bios?.version || "N/A"} (
                        {stats.systemInfo?.bios?.date || "N/A"})
                      </span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          <Card className="border-gray-700 bg-gray-900">
            <CardHeader className="pb-2">
              <div className="flex items-center gap-2">
                <Info className="h-4 w-4 text-green-400" />
                <CardTitle className="text-sm font-medium text-gray-200">
                  Software Version
                </CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="space-y-3">
                  <Skeleton className="h-5 w-full bg-gray-800" />
                  <Skeleton className="h-5 w-full bg-gray-800" />
                </div>
              ) : (
                <div className="space-y-3 text-sm">
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400 flex items-center gap-2">
                      <Pipe className="h-4 w-4" /> Version
                    </span>
                    {stats.systemInfo?.sysinfo?.version || stats.version}
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400 flex items-center gap-2">
                      <Clock className="h-4 w-4" /> Last Update
                    </span>
                    <span className="text-gray-300">{stats.lastUpdate}</span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="storage" className="mt-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card className="border-gray-700 bg-gray-900 md:col-span-1">
              <CardHeader className="pb-2">
                <div className="flex items-center gap-2">
                  <HardDrives className="h-4 w-4 text-amber-400" />
                  <CardTitle className="text-sm font-medium text-gray-200">
                    Storage Overview
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <Skeleton className="h-60 w-full bg-gray-800" />
                ) : (
                  <div className="flex flex-col items-center justify-center">
                    {storageData.length > 0 ? (
                      <div className="w-48 h-48 relative">
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={storageData}
                              cx="50%"
                              cy="50%"
                              labelLine={false}
                              outerRadius={80}
                              innerRadius={40}
                              fill="#8884d8"
                              dataKey="value"
                            >
                              {storageData.map((entry, index) => (
                                <Cell
                                  key={`cell-${index}`}
                                  fill={COLORS[index % COLORS.length]}
                                />
                              ))}
                            </Pie>
                          </PieChart>
                        </ResponsiveContainer>
                        <div className="absolute inset-0 flex items-center justify-center flex-col">
                          <div className="text-2xl font-bold text-gray-50">
                            {storageData.reduce(
                              (acc, disk) => acc + disk.value,
                              0
                            )}
                            GB
                          </div>
                          <div className="text-xs text-gray-400">
                            Total Storage
                          </div>
                        </div>
                      </div>
                    ) : (
                      <div className="text-gray-400 flex items-center justify-center h-60">
                        No storage data available
                      </div>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="border-gray-700 bg-gray-900 md:col-span-1">
              <CardHeader className="pb-2">
                <div className="flex items-center gap-2">
                  <HardDrives className="h-4 w-4 text-blue-400" />
                  <CardTitle className="text-sm font-medium text-gray-200">
                    Storage Devices
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <div className="space-y-3">
                    <Skeleton className="h-16 w-full bg-gray-800" />
                    <Skeleton className="h-16 w-full bg-gray-800" />
                    <Skeleton className="h-16 w-full bg-gray-800" />
                  </div>
                ) : (
                  <div className="space-y-3">
                    {stats.systemInfo?.storage?.map((disk, index) => (
                      <div
                        key={index}
                        className="bg-gray-800/50 p-3 rounded-lg border border-gray-700"
                      >
                        <div className="flex justify-between items-center">
                          <div className="flex items-center gap-2">
                            <div
                              className="w-3 h-3 rounded-full"
                              style={{
                                backgroundColor: COLORS[index % COLORS.length]
                              }}
                            ></div>
                            <span className="text-gray-200 font-medium">
                              {disk.name}
                            </span>
                          </div>
                          <Desktop className="bg-gray-800 text-gray-200 border-gray-700">
                            {disk.size}GB
                          </Desktop>
                        </div>
                        <div className="mt-2 text-xs text-gray-400">
                          <div className="flex justify-between">
                            <span>Model: {disk.model}</span>
                            <span>Serial: {disk.serial}</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="network" className="mt-4 space-y-4">
          <Card className="border-gray-700 bg-gray-900 md:col-span-1">
            <CardHeader className="pb-2">
              <div className="flex items-center gap-2">
                <Network className="h-4 w-4 text-green-400" />
                <CardTitle className="text-sm font-medium text-gray-200">
                  Network Status
                </CardTitle>
              </div>
              <CardDescription className="text-gray-400">
                Service health monitoring
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="grid grid-cols-2 gap-4">
                  <Skeleton className="h-10 w-full bg-gray-800" />
                  <Skeleton className="h-10 w-full bg-gray-800" />
                  <Skeleton className="h-10 w-full bg-gray-800" />
                  <Skeleton className="h-10 w-full bg-gray-800" />
                </div>
              ) : (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {stats.services.map((service, idx) => (
                    <div
                      key={idx}
                      className="bg-gray-800/50 p-3 rounded-lg border border-gray-700"
                    >
                      <div className="flex flex-col gap-2">
                        <span className="text-gray-300 text-sm font-medium">
                          {service.name}
                        </span>
                        <div className="flex justify-between items-center">
                          <span className="text-gray-400 text-xs">
                            :{service.port}
                          </span>
                          <p className={getStatusColor(service.status)}>
                            {service.status}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export interface Root {
  system: System;
}

export interface System {
  sysinfo: Sysinfo;
  node: Node;
  os: Os;
  kernel: Kernel;
  product: Product;
  board: Board;
  chassis: Chassis;
  bios: Bios;
  cpu: Cpu;
  memory: Memory;
  storage: Storage[];
  network: Network[];
}

export interface Sysinfo {
  version: string;
  timestamp: string;
}

export interface Node {
  hostname: string;
  machineid: string;
  timezone: string;
}

export interface Os {
  name: string;
  vendor: string;
  version: string;
  architecture: string;
}

export interface Kernel {
  release: string;
  version: string;
  architecture: string;
}

export interface Product {
  name: string;
  vendor: string;
  version: string;
  uuid: string;
  sku: string;
}

export interface Board {
  name: string;
  vendor: string;
  version: string;
  assettag: string;
}

export interface Chassis {
  type: number;
  vendor: string;
  version: string;
  assettag: string;
}

export interface Bios {
  vendor: string;
  version: string;
  date: string;
}

export interface Cpu {
  vendor: string;
  model: string;
  cache: number;
  cpus: number;
  cores: number;
  threads: number;
}

export interface Memory {}

export interface Storage {
  name: string;
  model: string;
  serial: string;
  size: number;
  driver?: string;
  vendor?: string;
}

export interface Network {
  name: string;
  driver: string;
  macaddress: string;
  port?: string;
  speed?: number;
}
