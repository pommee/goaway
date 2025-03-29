import { useState, useEffect } from "react";
import { GetRequest } from "@/util";
import { toast } from "sonner";
import { ClientCard } from "@/app/clients/card";

export type ClientEntry = {
  ip: string;
  lastSeen: string;
  mac: string;
  name: string;
  vendor: string;
};

export function Clients() {
  const [clients, setClients] = useState<ClientEntry[]>([]);

  useEffect(() => {
    async function fetchClients() {
      const [code, response] = await GetRequest("clients");
      if (code !== 200) {
        toast.warning(`Unable to fetch clients`);
        return;
      }

      if (Array.isArray(response.clients)) {
        setClients(response.clients);
      } else {
        console.warn("Unexpected response format:", response);
      }
    }

    fetchClients();
  }, []);

  return (
    <div>
      <div className="grid lg:grid-cols-3 gap-2">
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
    </div>
  );
}
