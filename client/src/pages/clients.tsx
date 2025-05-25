import { ClientCard } from "@/app/clients/card";
import { GetRequest } from "@/util";
import { Info } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export type ClientEntry = {
  ip: string;
  lastSeen: string;
  mac: string;
  name: string;
  vendor: string;
};

export function Clients() {
  const [clients, setClients] = useState<ClientEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchClients() {
      try {
        setLoading(true);
        const [code, response] = await GetRequest("clients");
        if (code !== 200) {
          toast.warning(response.error);
          return;
        }

        if (Array.isArray(response.clients)) {
          const clientsSorted = response.clients.sort(
            (a: { ip: number }, b: { ip: number }) => a.ip > b.ip
          );
          setClients(clientsSorted);
        } else {
          console.warn("Unexpected response format:", response);
        }
      } catch (error) {
        console.error("Error fetching clients:", error);
        toast.error("Failed to fetch clients");
      } finally {
        setLoading(false);
      }
    }

    fetchClients();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-pulse text-gray-500">Loading clients...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-full">
      {clients.length > 0 ? (
        <div className="grid lg:grid-cols-4 gap-2">
          {clients.map((client, index) => (
            <ClientCard
              key={index}
              ip={client.ip}
              lastSeen={client.lastSeen}
              mac={client.mac}
              name={client.name}
              vendor={client.vendor}
            />
          ))}
        </div>
      ) : (
        <div className="flex justify-center items-center py-32">
          <div className="border rounded-lg p-6 max-w-md w-full">
            <div className="flex items-center justify-center">
              <div className="p-3">
                <Info className="w-12 h-12 text-blue-400" />
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
