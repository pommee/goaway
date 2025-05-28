import { GetRequest } from "@/util";
import { useEffect, useRef, useState } from "react";
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
  type: "server" | "client";
  ip?: string;
  lastSeen?: string;
  mac?: string;
  vendor?: string;
  color?: string;
  size?: number;
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

export default function DNSServerVisualizer() {
  const [clients, setClients] = useState<ClientEntry[]>([]);
  const [selectedClient, setSelectedClient] = useState<ClientEntry | null>(
    null
  );
  const [selectedPosition, setSelectedPosition] = useState({ x: 0, y: 0 });
  const [networkData, setNetworkData] = useState<NetworkData>({
    nodes: [],
    links: []
  });
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [error, setError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions] = useState({
    width: window.innerWidth - 300,
    height: window.innerHeight - 300
  });
  const fgRef = useRef<ForceGraphMethods | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const createPulse = (
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
  };

  useEffect(() => {
    if (fgRef.current) {
      fgRef.current.d3Force("charge")?.strength(-20);
      fgRef.current.d3Force("link")?.distance(80);
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
          createPulse(communicationEvent.ip, "dns-server", "client");
        }

        if (communicationEvent.dns && communicationEvent.ip !== "") {
          createPulse("dns-server", communicationEvent.ip, "dns");
        } else if (communicationEvent.dns) {
          createPulse("dns-server", "upstream", "dns");
        }

        if (communicationEvent.upstream) {
          createPulse("upstream", "dns-server", "upstream");

          const matchingClient = clients.find(
            (client) => client.ip === communicationEvent.ip
          );

          if (matchingClient) {
            setTimeout(() => {
              createPulse("dns-server", communicationEvent.ip, "dns");
            }, 300);
          }
        }
      } catch (error) {
        console.error("Error handling WebSocket message:", error);
      }
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN) ws.close();
      wsRef.current = null;
    };
  }, [clients]);

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
    if (clients.length === 0) return;

    const nodes: NetworkNode[] = [
      {
        id: "dns-server",
        name: "DNS Server",
        type: "server",
        color: "#3b82f6",
        size: 8
      },
      {
        id: "upstream",
        name: "Upstream",
        type: "server",
        color: "#008000",
        size: 8
      }
    ];

    const links: NetworkLink[] = [];

    links.push({
      source: "upstream",
      target: "dns-server",
      color: "#313131",
      width: 1
    });

    clients.forEach((client) => {
      const nodeColor = "#ef4444";
      const linkColor = "#313131";

      nodes.push({
        id: client.ip,
        name: client.name || client.ip,
        type: "client",
        ip: client.ip,
        lastSeen: client.lastSeen,
        mac: client.mac,
        vendor: client.vendor,
        color: nodeColor,
        size: 5
      });

      links.push({
        source: client.ip,
        target: "dns-server",
        color: linkColor,
        width: 1
      });
    });

    setNetworkData({ nodes, links });
  }, [clients]);

  const handleNodeClick = (node: NetworkNode, event: MouseEvent) => {
    if (node.type === "client") {
      const client = clients.find((c) => c.ip === node.id);
      if (client) {
        setSelectedClient(client);
        setSelectedPosition({ x: event.clientX, y: event.clientY });
      }
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
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x1, y1);
      ctx.lineTo(x2, y2);
      ctx.stroke();
    });
  };

  if (error) {
    return (
      <div className="p-4 min-h-screen text-white">
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
    <div className="text-white">
      <div className="mb-4">
        <h1 className="text-2xl font-semibold">DNS Server Network Map</h1>
        <p className="text-sm text-muted-foreground">
          Live visualization of connected clients
        </p>
      </div>

      <div
        ref={containerRef}
        className="rounded-xl border border-stone-800 bg-stone-950 shadow-md px-2"
      >
        <div className="grid grid-cols-3 gap-2 text-sm mt-2">
          {[
            { label: "client", plural: "clients", value: clients.length },
            { label: "node", plural: "nodes", value: networkData.nodes.length },
            { label: "link", plural: "links", value: networkData.links.length }
          ].map(({ label, plural, value }) => (
            <div
              key={label}
              className="rounded-lg bg-stone-900/80 border border-stone-800 py-0.5 text-center shadow-sm"
            >
              <p className="text-sm font-medium text-white">{value}</p>
              <p className="text-xs text-muted-foreground">
                {value === 1 ? label : plural}
              </p>
            </div>
          ))}
        </div>

        <p className="flex mt-4 ml-1 text-muted-foreground text-sm">
          Client nodes can be clicked for more information.
        </p>
        <p className="flex ml-1 text-muted-foreground text-sm">
          Nodes can be dragged around.
        </p>

        {networkData.nodes.length > 0 && (
          <div className="rounded-md cursor-move">
            <ForceGraph2D
              ref={fgRef}
              graphData={networkData}
              width={dimensions.width}
              height={dimensions.height}
              nodeColor={(node: NetworkNode) => node.color || "#ffffff"}
              nodeVal={(node: NetworkNode) => node.size || 1}
              nodeLabel={(node: NetworkNode) => node.ip || ""}
              linkColor={(link: NetworkLink) => link.color || "#313131"}
              linkWidth={(link: NetworkLink) => link.width || 1}
              onNodeClick={handleNodeClick}
              nodeCanvasObjectMode={() => "after"}
              nodeCanvasObject={(
                node: NetworkNode & { x: number; y: number },
                ctx,
                globalScale
              ) => {
                const label = node.name;
                const fontSize = 12 / globalScale;
                ctx.font = `${fontSize}px Sans-Serif`;
                ctx.textAlign = "center";
                ctx.textBaseline = "middle";
                ctx.fillStyle = "white";
                ctx.fillText(
                  label,
                  node.x,
                  node.y + 2 + ((node.size || 5) + fontSize)
                );
              }}
              linkCanvasObjectMode={() => "replace"}
              linkCanvasObject={renderCustomLink}
              cooldownTicks={100}
              d3AlphaDecay={0.0228}
              d3VelocityDecay={0.4}
            />
          </div>
        )}

        <p className="text-right text-xs text-muted-foreground italic mb-2">
          use mouse to move and zoom
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
    </div>
  );
}
