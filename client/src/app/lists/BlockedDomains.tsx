import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DeleteRequest, GetRequest } from "@/util";
import { MagnifyingGlassIcon, XIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

async function removeDomain(domain: string) {
  const [status, response] = await DeleteRequest(
    `blacklist?domain=${domain}`,
    null
  );

  if (status === 200) {
    toast.success(`Removed ${domain}`);
    return true;
  } else {
    toast.error(
      response &&
      typeof response === "object" &&
      "error" in response &&
      typeof (response as { error?: unknown }).error === "string"
        ? (response as { error: string }).error
        : "Error removing domain"
    );
    return false;
  }
}

export default function BlockedDomainsList({ listName }: { listName: string }) {
  const [domains, setDomains] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);

  const deleteDomain = async (name: string) => {
    const success = await removeDomain(name);
    if (success) {
      setDomains(domains.filter((domain) => domain !== name));
    }
  };

  useEffect(() => {
    async function fetchBlockedDomains() {
      try {
        setLoading(true);
        const [code, response] = await GetRequest(
          `getDomainsForList?list=${listName}`
        );
        if (code !== 200) {
          toast.warning("Unable to fetch client details");
          return;
        }

        if (Array.isArray(response)) {
          setDomains(response as string[]);
        } else {
          setDomains([]);
        }
      } catch {
        toast.error(`Could not fetch domains for ${listName}`);
      } finally {
        setLoading(false);
      }
    }

    fetchBlockedDomains();
  }, [listName]);

  const filteredDomains = domains
    ? domains.filter((domain) =>
        domain.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : [];

  return (
    <div className="w-full p-4 bg-zinc-900 rounded-lg text-white">
      <h2 className="text-xl font-bold mb-4">Blocked Domains</h2>

      <div className="flex gap-2 mb-4 flex-col md:flex-row">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-2 top-2 h-4 w-4 text-zinc-400" />
          <Input
            placeholder="Search domains..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8 bg-zinc-800 border-zinc-700 text-white"
          />
        </div>
      </div>

      <div className="bg-zinc-800 rounded-lg overflow-hidden">
        <div className="grid grid-cols-12 p-3 bg-zinc-700 font-medium">
          <div className="col-span-10">Domain</div>
          <div className="col-span-2 text-right">Actions</div>
        </div>

        <div className="max-h-96 overflow-y-auto">
          {loading ? (
            <div className="p-4 text-center text-zinc-400">
              Loading domains...
            </div>
          ) : filteredDomains.length > 0 ? (
            filteredDomains.map((domain, index) => (
              <div
                key={index}
                className="grid grid-cols-12 p-3 border-b border-zinc-700 hover:bg-zinc-750"
              >
                <div className="col-span-10 truncate">{domain}</div>
                <div className="col-span-2 text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => deleteDomain(domain)}
                    className="text-red-400 hover:text-red-300 hover:bg-red-900/30"
                  >
                    <XIcon className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))
          ) : (
            <div className="p-4 text-center text-zinc-400">
              {searchTerm
                ? "No domains match your search"
                : "No domains in blocklist"}
            </div>
          )}
        </div>
      </div>

      <div className="mt-2 text-sm text-zinc-400">
        {filteredDomains.length}{" "}
        {filteredDomains.length === 1 ? "domain" : "domains"}{" "}
        {searchTerm && "matching your search"}
      </div>
    </div>
  );
}
