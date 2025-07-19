import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import {
  CheckCircleIcon,
  DatabaseIcon,
  GlobeIcon,
  InfoIcon,
  MagnifyingGlassIcon,
  NetworkIcon,
  PlusIcon,
  TrashIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

type ListEntry = {
  ip: string;
  domain: string;
};

async function CreateResolution(domain: string, ip: string) {
  const [code, response] = await PostRequest("resolution", {
    ip,
    domain
  });
  if (code === 200) {
    toast.success(`${domain} has been added!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

async function DeleteResolution(domain: string, ip: string) {
  const [code, response] = await DeleteRequest(
    `resolution?domain=${domain}&ip=${ip}`,
    null
  );
  if (code === 200) {
    toast.success(`${domain} was deleted!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

export function Resolution() {
  const [resolutions, setResolutions] = useState<ListEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [domainName, setDomainName] = useState("");
  const [ip, setIP] = useState("");
  const [searchTerm, setSearchTerm] = useState("");

  const fetchResolutions = async () => {
    setLoading(true);
    const [code, response] = await GetRequest("resolutions");
    if (code !== 200) {
      toast.error("Unable to fetch resolutions");
      setLoading(false);
      return;
    }

    const listArray: ListEntry[] = Object.entries(response || {}).map(
      ([, details]) => ({
        domain: details.domain,
        ip: details.ip
      })
    );

    setResolutions(listArray);
    setLoading(false);
  };

  useEffect(() => {
    fetchResolutions();
  }, []);

  const handleSave = async () => {
    if (!domainName || !ip) {
      toast.warning("Both domain and IP are required");
      return;
    }

    setSubmitting(true);
    const success = await CreateResolution(domainName, ip);
    if (success) {
      await fetchResolutions();
      setDomainName("");
      setIP("");
    }
    setSubmitting(false);
  };

  const handleDelete = async (domain: string, ip: string) => {
    const success = await DeleteResolution(domain, ip);
    if (success) {
      await fetchResolutions();
    }
  };

  const filteredResolutions = searchTerm
    ? resolutions.filter(
        (res) =>
          res.domain.toLowerCase().includes(searchTerm.toLowerCase()) ||
          res.ip.includes(searchTerm)
      )
    : resolutions;

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            Custom DNS Resolutions
          </h1>
          <p className="text-muted-foreground mt-1">
            Map custom domains to specific IP addresses
          </p>
        </div>
        <div className="flex items-center gap-2">
          <DatabaseIcon className="h-3 w-3" />
          {resolutions.length} {resolutions.length === 1 ? "Entry" : "Entries"}
        </div>
      </div>

      <Card className="shadow-md">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2">
            <PlusIcon className="h-5 w-5 text-green-500" />
            Add New Resolution
          </CardTitle>
          <CardDescription>
            Create a custom domain-to-IP mapping for your network
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 lg:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="domain" className="font-medium">
                Domain name
              </Label>
              <div className="relative">
                <GlobeIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="domain"
                  placeholder="example.local"
                  className="pl-9"
                  value={domainName}
                  onChange={(e) => setDomainName(e.target.value)}
                />
                <p className="text-sm text-muted-foreground mt-1">
                  Domain name to use, supports wildcard.
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="ip" className="font-medium">
                IP Address
              </Label>
              <Input
                id="ip"
                placeholder="192.168.1.100"
                value={ip}
                onChange={(e) => setIP(e.target.value)}
              />
              <p className="text-sm text-muted-foreground mt-1">
                IPv4 / IPv6 address where domains will resolve
              </p>
            </div>

            <div className="space-y-2">
              <Label className="font-medium text-transparent">Action</Label>
              <Button
                variant="default"
                className="w-full"
                onClick={handleSave}
                disabled={submitting || !domainName || !ip}
              >
                {submitting ? (
                  <>
                    <div className="h-4 w-4 mr-2 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    Saving...
                  </>
                ) : (
                  "Save Resolution"
                )}
              </Button>
            </div>
          </div>

          <div className="mt-4 bg-accent border rounded-xl p-3">
            <div className="flex items-start gap-3">
              <InfoIcon className="h-4 w-4 text-blue-400 mt-0.5 flex-shrink-0" />
              <div className="space-y-2 flex-1">
                <div>
                  <h4 className="font-medium mb-1">Wildcard Matching</h4>
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    Use wildcards to match multiple subdomains with a single
                    rule
                  </p>
                </div>

                <div className="bg-background rounded-lg p-3">
                  <div className="flex items-center justify-between mb-2">
                    <code className="text-sm font-mono bg-accent px-2 py-1 rounded text-blue-400 font-medium">
                      *.example.local
                    </code>
                    <div className="flex items-center gap-1.5 text-emerald-400">
                      <CheckCircleIcon className="h-3 w-3" />
                      <span className="text-xs font-medium">Matches</span>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-1">
                    {[
                      "app.example.local",
                      "my.app.example.local",
                      "sub1.sub2.sub3.example.local"
                    ].map((domain, index) => (
                      <div
                        key={index}
                        className="flex items-center gap-1.5 text-xs text-muted-foreground"
                      >
                        <div className="w-1 h-1 bg-accent-foreground rounded-full" />
                        <code className="font-mono truncate">{domain}</code>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="py-4">
        <CardHeader className="pb-4 border-b">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-3">
              <div className="bg-blue-500/20 p-2 rounded-lg">
                <DatabaseIcon className="h-5 w-5 text-blue-400" />
              </div>
              <div>
                <span>Current Resolutions</span>
                <p className="text-sm text-muted-foreground font-normal mt-0.5">
                  {resolutions.length} active{" "}
                  {resolutions.length === 1 ? "mapping" : "mappings"}
                </p>
              </div>
            </CardTitle>
            <div className="w-72">
              <div className="relative">
                <MagnifyingGlassIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search domains or IPs..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <div className="p-6 space-y-4">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="flex items-center justify-between p-4 rounded-lg border border-stone"
                >
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-24 bg-accent" />
                  </div>
                  <Skeleton className="h-8 w-8 rounded-full bg-accent" />
                </div>
              ))}
            </div>
          ) : filteredResolutions.length > 0 ? (
            <div className="divide-y divide-stone">
              {filteredResolutions.map((resolution) => (
                <div
                  key={resolution.domain}
                  className="group flex items-center justify-between p-2 hover:bg-accent transition-all duration-200"
                >
                  <div className="flex items-center gap-4 flex-1">
                    <div className="flex-shrink-0">
                      <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-1">
                        <GlobeIcon className="h-4 w-4 text-blue-400 flex-shrink-0" />
                        <span className="font-medium truncate">
                          {resolution.domain}
                        </span>
                        {resolution.domain.includes("*") && (
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-orange-200/20 text-orange-300 border border-orange-500/30">
                            Wildcard
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <NetworkIcon />
                        <code className="font-mono bg-accent px-2 py-0.5 rounded">
                          {resolution.ip}
                        </code>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/10 cursor-pointer"
                      onClick={() =>
                        handleDelete(resolution.domain, resolution.ip)
                      }
                    >
                      <TrashIcon className="h-4 w-4" />
                      <span className="sr-only">Delete</span>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-4 text-center">
              <div className="bg-accent p-4 rounded-full mb-4">
                <DatabaseIcon className="h-8 w-8" />
              </div>
              <h3 className="text-lg font-medium mb-2">No resolutions found</h3>
              <p className="text-muted-foreground max-w-sm">
                {searchTerm
                  ? "No matching entries for your search term. Try a different keyword."
                  : "Get started by adding your first custom DNS resolution above."}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
