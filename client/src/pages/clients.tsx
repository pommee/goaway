import DNSServerVisualizer from "@/app/clients/map";
import { NoContent } from "@/shared";
import { GetRequest } from "@/util";
import { SpinnerIcon } from "@phosphor-icons/react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

export type ClientEntry = {
  ip: string;
  lastSeen: string;
  name: string;
  mac: string;
  vendor: string;
  bypass: boolean;
  x?: number;
  y?: number;
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
        toast.warning(response);
        return;
      }

      if (Array.isArray(response)) {
        const clientsSorted = sortClientsByIP(response);
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
        <SpinnerIcon className="mr-2 animate-spin" />
        <div className="animate-pulse text-muted-foreground">
          Loading network map...
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-200px)]">
      {clients.length > 0 ? (
        <DNSServerVisualizer clients={clients} />
      ) : (
        <NoContent text="No clients have sent any requests yet. Client information will appear here when available." />
      )}
    </div>
  );
}
