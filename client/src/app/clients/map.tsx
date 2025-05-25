import { ClientEntry } from "@/pages/clients";
import { Info } from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CardDetails } from "./details";

interface DNSServerVisualizerProps {
  clients: ClientEntry[];
}

interface Pulse {
  id: number;
  startX: number;
  startY: number;
  endX: number;
  endY: number;
  progress: number;
  type: "request" | "response";
  clientName: string;
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

export function DNSServerVisualizer({ clients }: DNSServerVisualizerProps) {
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [stats, setStats] = useState({ requests: 0, responses: 0 });
  const pulseIdRef = useRef(0);
  const [wsConnected, setWsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const [selectedClient, setSelectedClient] = useState<ClientEntry | null>(
    null
  );
  const [windowWidth] = useState(window.innerWidth - 400);
  const [windowHeight] = useState(window.innerHeight - 300);

  const centerX = windowWidth / 2;
  const centerY = windowHeight / 2;
  const radius = 140;

  const serverPos = useMemo(
    () => ({ x: centerX, y: centerY }),
    [centerX, centerY]
  );
  const upstreamPos = useMemo(
    () => ({ x: centerX - 80, y: centerY - 200 }),
    [centerX, centerY]
  );

  const clientsWithPositions = useMemo(() => {
    return clients.map((client, index) => {
      const angle = (index * 360) / clients.length - 90;
      const angleRad = (angle * Math.PI) / 180;
      return {
        ...client,
        x: serverPos.x + radius * Math.cos(angleRad),
        y: serverPos.y + radius * Math.sin(angleRad)
      };
    });
  }, [clients, serverPos, radius]);

  const createPulse = useCallback(
    (
      startX: number,
      startY: number,
      endX: number,
      endY: number,
      type: "request" | "response",
      clientName: string
    ) => {
      const newPulse: Pulse = {
        id: pulseIdRef.current++,
        startX: startX,
        startY: startY,
        endX: endX,
        endY: endY,
        progress: 0,
        type: type,
        clientName: clientName
      };

      setPulses((prev) => [...prev, newPulse]);
      setStats((prev) => ({
        ...prev,
        [type === "request" ? "requests" : "responses"]:
          prev[type === "request" ? "requests" : "responses"] + 1
      }));
    },
    []
  );

  useEffect(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/liveCommunication`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => setWsConnected(true);
    ws.onerror = () => setWsConnected(false);
    ws.onclose = () => setWsConnected(false);

    ws.onmessage = (event) => {
      try {
        const communicationEvent: CommunicationEvent = JSON.parse(event.data);
        const matchingClient = clientsWithPositions.find(
          (client) => client.ip === communicationEvent.ip
        );

        if (communicationEvent.client) {
          if (matchingClient) {
            createPulse(
              matchingClient.x,
              matchingClient.y,
              serverPos.x,
              serverPos.y,
              "request",
              matchingClient.name || matchingClient.ip
            );
          }
        }

        if (communicationEvent.dns && communicationEvent.ip !== "") {
          createPulse(
            serverPos.x,
            serverPos.y,
            matchingClient.x,
            matchingClient.y,
            "response",
            "DNS Server"
          );
        } else if (communicationEvent.dns) {
          createPulse(
            serverPos.x,
            serverPos.y,
            upstreamPos.x,
            upstreamPos.y,
            "request",
            "DNS Server"
          );
        }

        if (communicationEvent.upstream) {
          createPulse(
            upstreamPos.x,
            upstreamPos.y,
            serverPos.x,
            serverPos.y,
            "response",
            "Upstream"
          );

          const matchingClient = clientsWithPositions.find(
            (client) => client.ip === communicationEvent.ip
          );

          if (matchingClient) {
            setTimeout(() => {
              createPulse(
                serverPos.x,
                serverPos.y,
                matchingClient.x,
                matchingClient.y,
                "response",
                "DNS Server"
              );
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
  }, [clientsWithPositions, createPulse, serverPos, upstreamPos]);

  useEffect(() => {
    const interval = setInterval(() => {
      setPulses((prev) => {
        return prev
          .map((pulse) => ({ ...pulse, progress: pulse.progress + 1 / 25 }))
          .filter((pulse) => pulse.progress < 1);
      });
    }, 16);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex flex-col ">
      <div className="p-4">
        <h1 className="text-xl font-bold text-white">
          DNS Server Communication Monitor
        </h1>
        <p className="text-muted-foreground mb-2">
          Visualizes live traffic between GoAway, upstream server and clients
        </p>
        <div className="flex gap-6 text-sm">
          <div className="flex items-center gap-2">
            <div
              className={`w-3 h-3 rounded-full ${
                wsConnected ? "bg-green-400 animate-pulse" : "bg-red-400"
              }`}
            ></div>
            <span className={wsConnected ? "text-green-400" : "text-red-400"}>
              WebSocket: {wsConnected ? "Connected" : "Disconnected"}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-blue-400 rounded-full animate-pulse"></div>
            <span className="text-blue-400">Requests: {stats.requests}</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
            <span className="text-green-400">Responses: {stats.responses}</span>
          </div>
        </div>
      </div>

      <div className="w-full flex-1 flex items-center justify-center p-4">
        <div className="rounded-2xl border border-white/10 p-4 shadow-2xl">
          <p className="text-sm text-muted-foreground flex">
            <Info className="mr-2 mt-0.5 text-blue-400" />
            Click client nodes to show more information
          </p>
          <svg width={windowWidth} height={windowHeight} className="block">
            <defs>
              {/* Request pulse gradients */}
              {pulses
                .filter((pulse) => pulse.type === "request")
                .map((pulse) => {
                  const gradientId = `request-gradient-${pulse.id}`;
                  const offset = pulse.progress * 100;

                  return (
                    <linearGradient
                      key={gradientId}
                      id={gradientId}
                      x1={pulse.startX}
                      y1={pulse.startY}
                      x2={pulse.endX}
                      y2={pulse.endY}
                      gradientUnits="userSpaceOnUse"
                    >
                      <stop
                        offset={`${Math.max(0, offset - 15)}%`}
                        stopColor="rgba(128, 128, 128, 0.2)"
                      />
                      <stop
                        offset={`${Math.max(0, offset - 8)}%`}
                        stopColor="rgba(59, 130, 246, 0.8)"
                      />
                      <stop
                        offset={`${offset}%`}
                        stopColor="rgba(59, 130, 246, 1)"
                      />
                      <stop
                        offset={`${Math.min(100, offset + 8)}%`}
                        stopColor="rgba(59, 130, 246, 0.8)"
                      />
                      <stop
                        offset={`${Math.min(100, offset + 15)}%`}
                        stopColor="rgba(128, 128, 128, 0.2)"
                      />
                    </linearGradient>
                  );
                })}

              {/* Response pulse gradients */}
              {pulses
                .filter((pulse) => pulse.type === "response")
                .map((pulse) => {
                  const gradientId = `response-gradient-${pulse.id}`;
                  const offset = pulse.progress * 100;

                  return (
                    <linearGradient
                      key={gradientId}
                      id={gradientId}
                      x1={pulse.startX}
                      y1={pulse.startY}
                      x2={pulse.endX}
                      y2={pulse.endY}
                      gradientUnits="userSpaceOnUse"
                    >
                      <stop
                        offset={`${Math.max(0, offset - 15)}%`}
                        stopColor="rgba(128, 128, 128, 0.2)"
                      />
                      <stop
                        offset={`${Math.max(0, offset - 8)}%`}
                        stopColor="rgba(34, 197, 94, 0.8)"
                      />
                      <stop
                        offset={`${offset}%`}
                        stopColor="rgba(34, 197, 94, 1)"
                      />
                      <stop
                        offset={`${Math.min(100, offset + 8)}%`}
                        stopColor="rgba(34, 197, 94, 0.8)"
                      />
                      <stop
                        offset={`${Math.min(100, offset + 15)}%`}
                        stopColor="rgba(128, 128, 128, 0.2)"
                      />
                    </linearGradient>
                  );
                })}
            </defs>

            {/* Static connection lines */}
            {clientsWithPositions.map((client, index) => (
              <line
                key={`static-${client.ip}-${index}`}
                x1={client.x}
                y1={client.y}
                x2={serverPos.x}
                y2={serverPos.y}
                stroke="rgba(255,255,255,0.2)"
                strokeWidth="1"
                strokeDasharray="4,2"
              />
            ))}

            {/* Server to upstream line */}
            <line
              x1={serverPos.x}
              y1={serverPos.y}
              x2={upstreamPos.x}
              y2={upstreamPos.y}
              stroke="rgba(255,255,255,0.3)"
              strokeWidth="2"
              strokeDasharray="6,3"
            />

            {/* Animated pulse lines */}
            {pulses.map((pulse) => {
              const gradientId = `${pulse.type}-gradient-${pulse.id}`;
              return (
                <line
                  key={`pulse-${pulse.id}`}
                  x1={pulse.startX}
                  y1={pulse.startY}
                  x2={pulse.endX}
                  y2={pulse.endY}
                  stroke={`url(#${gradientId})`}
                  strokeWidth="3"
                  strokeLinecap="round"
                />
              );
            })}

            {/* Client nodes */}
            {clientsWithPositions.map((client, index) => {
              const bubbleRadius =
                clients.length > 20 ? 18 : clients.length > 10 ? 22 : 25;
              const fontSize = clients.length > 20 ? "10px" : "12px";

              return (
                <g key={`${client.ip}-${index}`}>
                  <circle
                    onClick={() => setSelectedClient(client)}
                    cx={client.x}
                    cy={client.y}
                    r={bubbleRadius}
                    fill="rgba(59, 130, 246, 0.9)"
                    className="cursor-pointer hover:fill-opacity-100 transition-all"
                    stroke="rgba(255,255,255,0.3)"
                    strokeWidth="1"
                  />
                  <text
                    onClick={() => setSelectedClient(client)}
                    x={client.x}
                    y={client.y + 2}
                    textAnchor="middle"
                    className="fill-white font-medium cursor-pointer pointer-events-none"
                    style={{ fontSize }}
                  >
                    {client.name
                      ? client.name.substring(0, clients.length > 20 ? 4 : 6)
                      : client.ip.split(".").pop()}
                  </text>
                  {clients.length <= 15 && (
                    <text
                      onClick={() => setSelectedClient(client)}
                      x={client.x}
                      y={client.y + bubbleRadius + 12}
                      textAnchor="middle"
                      className="fill-white/70 cursor-pointer pointer-events-none"
                      style={{ fontSize: "10px" }}
                    >
                      {client.ip}
                    </text>
                  )}
                </g>
              );
            })}

            {/* DNS Server */}
            <g>
              <circle
                cx={serverPos.x}
                cy={serverPos.y}
                r="35"
                fill="rgba(249, 115, 22, 0.9)"
              />
              <text
                x={serverPos.x}
                y={serverPos.y + 3}
                textAnchor="middle"
                className="fill-white text-sm font-bold"
              >
                GoAway
              </text>
            </g>

            {/* Upstream */}
            <g>
              <circle
                cx={upstreamPos.x}
                cy={upstreamPos.y}
                r="35"
                fill="rgba(0, 180, 0, 0.8)"
              />
              <text
                x={upstreamPos.x}
                y={upstreamPos.y + 3}
                textAnchor="middle"
                className="fill-white text-sm font-bold"
              >
                Upstream
              </text>
            </g>
          </svg>
        </div>
      </div>

      {selectedClient && (
        <CardDetails
          key={selectedClient.ip}
          ip={selectedClient.ip}
          lastSeen={timeAgo(selectedClient.lastSeen)}
          mac={selectedClient.mac}
          name={selectedClient.name || ""}
          vendor={selectedClient.vendor}
          x={0}
          y={0}
          onClose={() => setSelectedClient(null)}
        />
      )}
    </div>
  );
}

export default DNSServerVisualizer;
