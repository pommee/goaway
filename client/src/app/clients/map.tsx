import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Slider } from "@/components/ui/slider";
import { GetRequest } from "@/util";
import { XIcon } from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import ForceGraph2D, { ForceGraphMethods } from "react-force-graph-2d";
import { CardDetails } from "./details";

interface ClientEntry {
  ip: string;
  lastSeen: string;
  mac: string;
  name: string;
  vendor: string;
}

interface NetworkNode {
  id: string;
  name: string;
  type: "server" | "client" | "cluster";
  ip?: string;
  lastSeen?: string;
  mac?: string;
  vendor?: string;
  color?: string;
  size?: number;
  clients?: ClientEntry[];
  subnet?: string;
  isActive?: boolean;
}

interface NetworkLink {
  source: string;
  target: string;
  color?: string;
  width?: number;
}

interface NetworkData {
  nodes: NetworkNode[];
  links: NetworkLink[];
}

interface Pulse {
  id: string;
  sourceId: string;
  targetId: string;
  progress: number;
  color: string;
  type: "client" | "dns" | "upstream";
}

interface CommunicationEvent {
  client: boolean;
  upstream: boolean;
  dns: boolean;
  ip: string;
}

interface ViewSettings {
  clusterBySubnet: boolean;
  hideInactiveClients: boolean;
  minNodeSize: number;
  maxNodeSize: number;
  showLabels: boolean;
  activityThresholdMinutes: number;
}

function timeAgo(timestamp: string) {
  const now = new Date();
  const past = new Date(timestamp);
  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000);

  const seconds = diffInSeconds % 60;
  const minutes = Math.floor((diffInSeconds / 60) % 60);
  const hours = Math.floor(diffInSeconds / 3600);

  return hours > 0
    ? `${hours}h ${minutes}m ${seconds}s ago`
    : `${minutes}m ${seconds}s ago`;
}

function getSubnet(ip: string): string {
  const parts = ip.split(".");
  return `${parts[0]}.${parts[1]}.${parts[2]}.x`;
}

function isClientActive(lastSeen: string, thresholdMinutes: number): boolean {
  const now = new Date();
  const past = new Date(lastSeen);
  const diffInMinutes = (now.getTime() - past.getTime()) / (1000 * 60);
  return diffInMinutes <= thresholdMinutes;
}

function groupClientsBySubnet(
  clients: ClientEntry[]
): Map<string, ClientEntry[]> {
  const subnetGroups = new Map<string, ClientEntry[]>();

  clients.forEach((client) => {
    const subnet = getSubnet(client.ip);
    if (!subnetGroups.has(subnet)) {
      subnetGroups.set(subnet, []);
    }
    subnetGroups.get(subnet)!.push(client);
  });

  return subnetGroups;
}

export default function DNSServerVisualizer() {
  const [clients, setClients] = useState<ClientEntry[]>([]);
  const [selectedClient, setSelectedClient] = useState<ClientEntry | null>(
    null
  );
  const [selectedCluster, setSelectedCluster] = useState<ClientEntry[] | null>(
    null
  );
  const [selectedPosition, setSelectedPosition] = useState({ x: 0, y: 0 });
  const [networkData, setNetworkData] = useState<NetworkData>({
    nodes: [],
    links: []
  });
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [viewSettings, setViewSettings] = useState<ViewSettings>({
    clusterBySubnet: false,
    hideInactiveClients: false,
    minNodeSize: 3,
    maxNodeSize: 12,
    showLabels: true,
    activityThresholdMinutes: 60
  });

  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({
    width: 800,
    height: 600
  });
  const fgRef = useRef<ForceGraphMethods | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const createPulse = useCallback(
    (
      sourceId: string,
      targetId: string,
      type: "client" | "dns" | "upstream"
    ) => {
      const colors = {
        client: "#22c55e",
        dns: "#3b82f6",
        upstream: "#f59e0b"
      };

      const newPulse: Pulse = {
        id: `${sourceId}-${targetId}-${Date.now()}-${Math.random()}`,
        sourceId,
        targetId,
        progress: 0,
        color: colors[type],
        type
      };

      setPulses((prev) => [...prev, newPulse]);
    },
    []
  );

  const filteredClients = useMemo(() => {
    return clients.filter((client) => {
      const matchesSearch =
        searchTerm === "" ||
        client.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        client.ip.includes(searchTerm) ||
        client.vendor.toLowerCase().includes(searchTerm.toLowerCase());

      const isActive = isClientActive(
        client.lastSeen,
        viewSettings.activityThresholdMinutes
      );
      const showInactive = !viewSettings.hideInactiveClients || isActive;

      return matchesSearch && showInactive;
    });
  }, [
    clients,
    searchTerm,
    viewSettings.hideInactiveClients,
    viewSettings.activityThresholdMinutes
  ]);

  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        setDimensions({
          width: rect.width - 32,
          height: Math.max(600, window.innerHeight - 400)
        });
      }
    };

    updateDimensions();
    window.addEventListener("resize", updateDimensions);

    return () => window.removeEventListener("resize", updateDimensions);
  }, []);

  useEffect(() => {
    if (fgRef.current) {
      const nodeCount = networkData.nodes.length;
      const chargeStrength = Math.max(-50, -300 / Math.sqrt(nodeCount));
      const linkDistance = nodeCount > 20 ? 120 : 80;

      fgRef.current.d3Force("charge")?.strength(chargeStrength);
      fgRef.current.d3Force("link")?.distance(linkDistance);
    }
  }, [networkData]);

  useEffect(() => {
    const fetchClients = async () => {
      try {
        setError(null);
        const [code, response] = await GetRequest("clients");

        if (code !== 200) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        setClients(response.clients);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch clients"
        );
        console.error("Error fetching clients:", err);
      }
    };

    fetchClients();
  }, []);

  useEffect(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/liveCommunication`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const communicationEvent: CommunicationEvent = JSON.parse(event.data);

        if (communicationEvent.client) {
          const sourceId = viewSettings.clusterBySubnet
            ? getSubnet(communicationEvent.ip)
            : communicationEvent.ip;
          createPulse(sourceId, "dns-server", "client");
        }

        if (communicationEvent.dns && communicationEvent.ip !== "") {
          const targetId = viewSettings.clusterBySubnet
            ? getSubnet(communicationEvent.ip)
            : communicationEvent.ip;
          createPulse("dns-server", targetId, "dns");
        } else if (communicationEvent.dns) {
          createPulse("dns-server", "upstream", "dns");
        }

        if (communicationEvent.upstream) {
          createPulse("upstream", "dns-server", "upstream");

          setClients((currentClients) => {
            const matchingClient = currentClients.find(
              (client) => client.ip === communicationEvent.ip
            );

            if (matchingClient) {
              setTimeout(() => {
                const targetId = viewSettings.clusterBySubnet
                  ? getSubnet(communicationEvent.ip)
                  : communicationEvent.ip;
                createPulse("dns-server", targetId, "dns");
              }, 300);
            }

            return currentClients;
          });
        }
      } catch (error) {
        console.error("Error handling WebSocket message:", error);
      }
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN) ws.close();
      wsRef.current = null;
    };
  }, [viewSettings.clusterBySubnet]);

  useEffect(() => {
    const interval = setInterval(() => {
      setPulses((prev) => {
        return prev
          .map((pulse) => ({ ...pulse, progress: pulse.progress + 1 / 12 }))
          .filter((pulse) => pulse.progress < 1);
      });
    }, 16);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (filteredClients.length === 0) return;

    const nodes: NetworkNode[] = [
      {
        id: "dns-server",
        name: "DNS Server",
        type: "server",
        color: "#3b82f6",
        size: viewSettings.maxNodeSize
      },
      {
        id: "upstream",
        name: "Upstream",
        type: "server",
        color: "#008000",
        size: viewSettings.maxNodeSize
      }
    ];

    const links: NetworkLink[] = [];

    links.push({
      source: "upstream",
      target: "dns-server",
      color: "#313131",
      width: 2
    });

    if (viewSettings.clusterBySubnet && filteredClients.length > 10) {
      const subnetGroups = groupClientsBySubnet(filteredClients);

      subnetGroups.forEach((subnetClients, subnet) => {
        const activeCount = subnetClients.filter((c) =>
          isClientActive(c.lastSeen, viewSettings.activityThresholdMinutes)
        ).length;

        const nodeSize = Math.max(
          viewSettings.minNodeSize,
          Math.min(viewSettings.maxNodeSize, 4 + subnetClients.length * 0.5)
        );

        const nodeColor = activeCount > 0 ? "#ef4444" : "#6b7280";

        nodes.push({
          id: subnet,
          name: `${subnet} (${subnetClients.length})`,
          type: "cluster",
          color: nodeColor,
          size: nodeSize,
          clients: subnetClients,
          subnet: subnet,
          isActive: activeCount > 0
        });

        links.push({
          source: subnet,
          target: "dns-server",
          color: "#313131",
          width: Math.max(1, Math.min(3, subnetClients.length * 0.1))
        });
      });
    } else {
      filteredClients.forEach((client) => {
        const isActive = isClientActive(
          client.lastSeen,
          viewSettings.activityThresholdMinutes
        );
        const nodeColor = isActive ? "#008000" : "#6b7280";
        const linkColor = isActive ? "#002000" : "#4b5563";

        nodes.push({
          id: client.ip,
          name: client.name || client.ip,
          type: "client",
          ip: client.ip,
          lastSeen: client.lastSeen,
          mac: client.mac,
          vendor: client.vendor,
          color: nodeColor,
          size: isActive
            ? viewSettings.maxNodeSize * 0.6
            : viewSettings.minNodeSize,
          isActive
        });

        links.push({
          source: client.ip,
          target: "dns-server",
          color: linkColor,
          width: isActive ? 1.5 : 0.5
        });
      });
    }

    setNetworkData({ nodes, links });
  }, [filteredClients, viewSettings]);

  const handleNodeClick = (node: NetworkNode, event: MouseEvent) => {
    if (node.type === "client") {
      const client = clients.find((c) => c.ip === node.id);
      if (client) {
        setSelectedClient(client);
        setSelectedCluster(null);
        setSelectedPosition({ x: event.clientX, y: event.clientY });
      }
    } else if (node.type === "cluster" && node.clients) {
      setSelectedCluster(node.clients);
      setSelectedClient(null);
      setSelectedPosition({ x: event.clientX, y: event.clientY });
    }
  };

  const renderCustomLink = (
    link: NetworkLink & {
      source: NetworkNode & { x: number; y: number };
      target: NetworkNode & { x: number; y: number };
    },
    ctx: CanvasRenderingContext2D
  ) => {
    const { source, target } = link;

    ctx.strokeStyle = link.color || "#313131";
    ctx.lineWidth = link.width || 0.5;
    ctx.beginPath();
    ctx.moveTo(source.x, source.y);
    ctx.lineTo(target.x, target.y);
    ctx.stroke();

    const linkPulses = pulses.filter(
      (pulse) =>
        (pulse.sourceId === source.id && pulse.targetId === target.id) ||
        (pulse.sourceId === target.id && pulse.targetId === source.id)
    );

    linkPulses.forEach((pulse) => {
      const isReverse =
        pulse.sourceId === target.id && pulse.targetId === source.id;
      const progress = isReverse ? 1 - pulse.progress : pulse.progress;

      const x1 = source.x + (target.x - source.x) * (progress - 0.1);
      const y1 = source.y + (target.y - source.y) * (progress - 0.1);
      const x2 = source.x + (target.x - source.x) * (progress + 0.1);
      const y2 = source.y + (target.y - source.y) * (progress + 0.1);

      const grad = ctx.createLinearGradient(x1, y1, x2, y2);
      grad.addColorStop(0, pulse.color + "00");
      grad.addColorStop(0.5, pulse.color);
      grad.addColorStop(1, pulse.color + "00");

      ctx.strokeStyle = grad;
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.moveTo(x1, y1);
      ctx.lineTo(x2, y2);
      ctx.stroke();
    });
  };

  if (error) {
    return (
      <div className="p-4 min-h-screen">
        <div className="text-center">
          <h1 className="text-xl font-bold mb-4">DNS Server Network Map</h1>
          <div className="bg-red-900/20 border border-red-500 rounded-lg p-4 max-w-md mx-auto">
            <p className="text-red-400 mb-2">Failed to connect to API:</p>
            <p className="text-sm text-gray-300">{error}</p>
            <p className="text-xs text-gray-400 mt-2">
              Could not load network map!
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-2 p-2 rounded-lg">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">
              Search Clients
            </label>
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search by name, IP, or vendor..."
            />
          </div>

          <div className="space-y-2 pt-2">
            <label className="flex items-center">
              <Checkbox
                checked={viewSettings.clusterBySubnet}
                onCheckedChange={(checked: boolean) =>
                  setViewSettings((prev) => ({
                    ...prev,
                    clusterBySubnet: checked
                  }))
                }
                className="mr-2"
              />
              <span className="text-sm">Cluster by subnet</span>
            </label>

            <label className="flex items-center">
              <Checkbox
                checked={viewSettings.hideInactiveClients}
                onCheckedChange={(checked: boolean) =>
                  setViewSettings((prev) => ({
                    ...prev,
                    hideInactiveClients: checked
                  }))
                }
                className="mr-2"
              />
              <span className="text-sm">Hide inactive clients</span>
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Activity threshold (minutes)
            </label>
            <Slider
              className="mb-1"
              defaultValue={[viewSettings.activityThresholdMinutes]}
              max={480}
              step={5}
              onValueChange={(newValue) =>
                setViewSettings((prev) => ({
                  ...prev,
                  activityThresholdMinutes: newValue[0]
                }))
              }
            />
            <div className="text-xs text-muted-foreground">
              {viewSettings.activityThresholdMinutes} minutes
            </div>
          </div>
        </div>
      </div>

      <div
        ref={containerRef}
        className="rounded-xl shadow-md p-4 w-full border-1 dark:bg-accent"
      >
        <div className="grid grid-cols-4 gap-2 text-sm mb-4">
          {[
            {
              label: "total client",
              plural: "total clients",
              value: clients.length
            },
            {
              label: "visible client",
              plural: "visible clients",
              value: filteredClients.length
            },
            { label: "node", plural: "nodes", value: networkData.nodes.length },
            {
              label: "active",
              plural: "active",
              value: filteredClients.filter((c) =>
                isClientActive(
                  c.lastSeen,
                  viewSettings.activityThresholdMinutes
                )
              ).length
            }
          ].map(({ label, plural, value }) => (
            <div key={label} className="rounded-lg py-0.5 text-center border-1">
              <p className="text-sm font-medium">{value}</p>
              <p className="text-xs text-muted-foreground">
                {value === 1 ? label : plural}
              </p>
            </div>
          ))}
        </div>

        <div className="mb-4 text-sm text-muted-foreground space-y-1">
          <p>Click nodes for details • Drag to move • Mouse wheel to zoom</p>
          <p>
            Green nodes are active • Gray nodes are inactive • Larger nodes =
            more clients
          </p>
        </div>

        {networkData.nodes.length > 0 && (
          <div className="rounded-md cursor-move overflow-hidden">
            <ForceGraph2D
              ref={fgRef}
              graphData={networkData}
              width={dimensions.width}
              height={dimensions.height}
              nodeColor={(node: NetworkNode) => node.color || "#ffffff"}
              nodeVal={(node: NetworkNode) => node.size || 1}
              nodeLabel={(node: NetworkNode) => {
                if (node.type === "cluster") {
                  return `${node.name}\nActive: ${
                    node.clients?.filter((c) =>
                      isClientActive(
                        c.lastSeen,
                        viewSettings.activityThresholdMinutes
                      )
                    ).length || 0
                  }`;
                }
                return node.ip || node.name || "";
              }}
              linkColor={(link: NetworkLink) => link.color || "#313131"}
              linkWidth={(link: NetworkLink) => link.width || 1}
              onNodeClick={handleNodeClick}
              nodeCanvasObjectMode={() =>
                viewSettings.showLabels ? "after" : "before"
              }
              nodeCanvasObject={
                viewSettings.showLabels
                  ? (
                      node: NetworkNode & { x: number; y: number },
                      ctx,
                      globalScale
                    ) => {
                      const label =
                        node.type === "cluster"
                          ? `${node.subnet} (${node.clients?.length})`
                          : node.name;
                      const fontSize = Math.max(8, 12 / globalScale);
                      ctx.font = `${fontSize}px Sans-Serif`;
                      ctx.textAlign = "center";
                      ctx.textBaseline = "middle";
                      ctx.fillStyle = node.isActive === false ? "#9ca3af" : "";
                      ctx.fillText(
                        label,
                        node.x,
                        node.y + 2 + ((node.size || 5) + fontSize)
                      );
                    }
                  : undefined
              }
              linkCanvasObjectMode={() => "replace"}
              linkCanvasObject={renderCustomLink}
              cooldownTicks={100}
              d3AlphaDecay={0.0228}
              d3VelocityDecay={0.4}
            />
          </div>
        )}

        <p className="text-right text-xs text-muted-foreground italic mt-2">
          Showing {filteredClients.length} of {clients.length} clients
        </p>
      </div>

      {selectedClient && (
        <CardDetails
          key={selectedClient.ip}
          ip={selectedClient.ip}
          lastSeen={timeAgo(selectedClient.lastSeen)}
          mac={selectedClient.mac}
          name={selectedClient.name}
          vendor={selectedClient.vendor}
          x={selectedPosition.x}
          y={selectedPosition.y}
          onClose={() => setSelectedClient(null)}
        />
      )}

      {selectedCluster && (
        <div
          className="fixed z-50 bg-stone-900 border rounded-lg p-4 shadow-lg max-w-md max-h-96 overflow-y-auto"
          style={{
            left: selectedPosition.x + 10,
            top: selectedPosition.y + 10
          }}
        >
          <ScrollArea>
            <div className="flex">
              <h3 className="font-semibold">
                Subnet Cluster ({selectedCluster.length} clients)
              </h3>
              <XIcon
                className="mt-0.5 ml-2 text-xl text-red-500 cursor-pointer"
                onClick={() => setSelectedCluster(null)}
              />
            </div>
          </ScrollArea>
          <div className="space-y-1 mt-2 cursor-pointer">
            {selectedCluster
              .sort((a, b) => new Date(b.lastSeen) - new Date(a.lastSeen))
              .map((client) => (
                <div
                  onClick={() => {
                    setSelectedClient(client);
                  }}
                  key={client.ip}
                  className={`p-2 rounded border text-sm ${
                    isClientActive(
                      client.lastSeen,
                      viewSettings.activityThresholdMinutes
                    )
                      ? "border-green-500 bg-green-900/20"
                      : "border-gray-600 bg-gray-800/20"
                  }`}
                >
                  <div className="font-medium">{client.name || client.ip}</div>
                  <div className="text-xs text-gray-400">
                    {client.ip} • {timeAgo(client.lastSeen)}
                  </div>
                </div>
              ))}
          </div>
        </div>
      )}
    </div>
  );
}
