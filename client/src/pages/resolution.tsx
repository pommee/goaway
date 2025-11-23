import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { NoContent } from "@/shared";
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import {
  CheckCircleIcon,
  DatabaseIcon,
  GlobeIcon,
  MagnifyingGlassIcon,
  NetworkIcon,
  PlusIcon,
  TrashIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { validateFQDN } from "@/pages/validation";

type ListEntry = {
  ip: string;
  domain: string;
};

async function CreateResolution(domain: string, ip: string) {
  const [code, response] = await PostRequest("resolution", { ip, domain });
  if (code === 200) {
    toast.success(`${domain} has been added!`);
    return true;
  } else {
    toast.error((response as { error?: string }).error || "Unknown error");
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
    toast.error((response as { error?: string }).error || "Unknown error");
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
  const [domainError, setDomainError] = useState<string | undefined>();

  useEffect(() => {
    (async () => {
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
    })();
  }, []);

  const handleDomainChange = (value: string) => {
    setDomainName(value);
    if (value.trim()) {
      const validation = validateFQDN(value);
      setDomainError(validation.error);
    } else {
      setDomainError(undefined);
    }
  };

  const handleSave = async () => {
    if (!domainName || !ip) {
      toast.warning("Both domain and IP are required");
      return;
    }

    const validation = validateFQDN(domainName);
    if (!validation.isValid) {
      toast.error(validation.error || "Invalid domain name");
      setDomainError(validation.error);
      return;
    }

    setSubmitting(true);
    const success = await CreateResolution(domainName, ip);
    if (success) {
      setResolutions((prev) => [...prev, { domain: domainName, ip }]);
      setDomainName("");
      setIP("");
      setDomainError(undefined);
    }
    setSubmitting(false);
  };

  const handleDelete = async (domain: string, ip: string) => {
    const success = await DeleteResolution(domain, ip);
    if (success) {
      setResolutions((prev) =>
        prev.filter((res) => !(res.domain === domain && res.ip === ip))
      );
    }
  };

  const filteredResolutions = searchTerm
    ? resolutions.filter(
        (res) =>
          res.domain.toLowerCase().includes(searchTerm.toLowerCase()) ||
          res.ip.includes(searchTerm)
      )
    : resolutions;

  const isFormValid = domainName && ip && !domainError;

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

      <div className="grid grid-cols-2 gap-5">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PlusIcon className="h-5 w-5 text-primary" />
              Add New Resolution
            </CardTitle>
            <CardDescription>
              Create a custom domain-to-IP mapping for your network
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4">
              <div>
                <div className="relative">
                  <GlobeIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="domain"
                    placeholder="example.local."
                    className={`pl-9 ${domainError ? "border-red-500" : ""}`}
                    value={domainName}
                    onChange={(e) => handleDomainChange(e.target.value)}
                  />
                  {domainError ? (
                    <p className="text-sm text-red-500 mt-1">{domainError}</p>
                  ) : (
                    <p className="text-sm text-muted-foreground mt-1">
                      Domain name to use, supports wildcard. Make sure it's a{" "}
                      <a
                        href="https://en.wikipedia.org/wiki/Fully_qualified_domain_name"
                        target="_blank"
                        rel="noreferrer"
                        className="underline hover:text-primary"
                      >
                        FQDN
                      </a>
                      .
                    </p>
                  )}
                </div>
              </div>

              <div>
                <Input
                  id="ip"
                  placeholder="192.168.1.100"
                  value={ip}
                  onChange={(e) => setIP(e.target.value)}
                />
                <p className="text-sm text-muted-foreground mt-1">
                  IPv4 / IPv6 address where domains will resolve.
                </p>
              </div>

              <div>
                <Button
                  variant="default"
                  className="w-full"
                  onClick={handleSave}
                  disabled={submitting || !isFormValid}
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
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Wildcard Matching</CardTitle>
            <CardDescription>
              Use wildcards to match multiple subdomains with a single rule
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-primary rounded-full" />
                  <span className="text-sm font-medium">Pattern</span>
                </div>
              </div>

              <code className="block px-3 py-2 rounded-md font-semibold border border-primary/50">
                *.example.local.
              </code>
            </div>

            <div className="space-y-2">
              <div className="grid gap-2">
                {["app.example.local.", "my.app.example.local."].map(
                  (domain, index) => (
                    <div
                      key={index}
                      className="flex items-center gap-3 p-2 rounded-md border"
                    >
                      <div className="w-1.5 h-1.5 bg-primary rounded-full" />
                      <code className="text-sm font-mono truncate flex-1">
                        {domain}
                      </code>
                      <CheckCircleIcon className="h-3 w-3 text-primary shrink-0" />
                    </div>
                  )
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

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
                  key={`${resolution.domain}-${resolution.ip}`}
                  className="group flex items-center justify-between p-2 hover:bg-accent transition-all duration-200"
                >
                  <div className="flex items-center gap-4 flex-1">
                    <div className="shrink-0">
                      <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-1">
                        <GlobeIcon className="h-4 w-4 text-blue-400 shrink-0" />
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
              <p className="text-muted-foreground">
                {searchTerm ? (
                  "No matching entries for your search term. Try a different keyword."
                ) : (
                  <NoContent text="Get started by adding your first custom DNS resolution above" />
                )}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
