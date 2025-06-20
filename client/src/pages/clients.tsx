import DNSServerVisualizer from "@/app/clients/map";
import { GetRequest } from "@/util";
import { InfoIcon } from "@phosphor-icons/react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

export type ClientEntry = {
  ip: string;
  lastSeen: string;
  mac: string;
  name: string;
  vendor: string;
  x: number;
  y: number;
};

const sortClientsByIP = (clients: ClientEntry[]): ClientEntry[] => {
  return [...clients].sort((a, b) => {
    const aNum = a.ip.split(".").map((num) => parseInt(num, 10));
    const bNum = b.ip.split(".").map((num) => parseInt(num, 10));

    for (let i = 0; i < 4; i++) {
      if (aNum[i] !== bNum[i]) {
        return aNum[i] - bNum[i];
      }
    }
    return 0;
  });
};

export function Clients() {
  const [clients, setClients] = useState<ClientEntry[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchClients = useCallback(async () => {
    try {
      setLoading(true);
      const [code, response] = await GetRequest("clients");

      if (code !== 200) {
        toast.warning(response.error);
        return;
      }

      if (Array.isArray(response.clients)) {
        const clientsSorted = sortClientsByIP(response.clients);
        setClients(clientsSorted);
      } else {
        console.warn("Unexpected response format:", response);
        setClients([]);
      }
    } catch (error) {
      console.error("Error fetching clients:", error);
      toast.error("Failed to fetch clients");
      setClients([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchClients();
  }, [fetchClients]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-pulse text-gray-500">
          Loading network map...
        </div>
      </div>
    );
  }

  return (
    <div>
      {clients.length > 0 ? (
        <DNSServerVisualizer clients={clients} />
      ) : (
        <div className="flex justify-center items-center py-32">
          <div className="border rounded-lg p-6 max-w-md w-full">
            <div className="flex items-center justify-center">
              <div className="p-3">
                <InfoIcon className="w-12 h-12 text-blue-400" />
              </div>
            </div>
            <h3 className="text-lg font-medium text-center">
              No Client Requests
            </h3>
            <p className="mt-2 text-center text-gray-600">
              No clients have sent any requests yet. New client information will
              appear here when available.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
