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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import {
  ClockIcon,
  DatabaseIcon,
  GlobeIcon,
  PlusIcon,
  SpinnerIcon,
  TrashIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

type PrefetchEntry = {
  domain: string;
  refresh: number;
  qtype: number;
};

function qtypeExpanded(qtype: number) {
  switch (qtype) {
    case 1:
      return "A";
    case 28:
      return "AAAA";
    case 5:
      return "CNAME";
    case 12:
      return "PTR";
  }
}

function validateFQDN(domain: string): { isValid: boolean; error?: string } {
  if (!domain || domain.trim() === "") {
    return { isValid: false, error: "Domain is required" };
  }

  const trimmedDomain = domain.trim();
  if (!trimmedDomain.endsWith(".")) {
    return {
      isValid: false,
      error:
        "Domain must end with a dot (.) to be a fully qualified domain name"
    };
  }

  const domainWithoutDot = trimmedDomain.slice(0, -1);
  if (domainWithoutDot.length === 0) {
    return { isValid: false, error: "Domain cannot be just a dot" };
  }

  if (domainWithoutDot.length > 253) {
    return {
      isValid: false,
      error: "Domain name is too long (max 253 characters)"
    };
  }

  // Check for valid characters and structure
  const domainRegex =
    /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/;

  if (!domainRegex.test(domainWithoutDot)) {
    return {
      isValid: false,
      error:
        "Invalid domain format. Use only letters, numbers, dots, and hyphens"
    };
  }

  const labels = domainWithoutDot.split(".");
  for (const label of labels) {
    if (label.length > 63) {
      return {
        isValid: false,
        error: "Each part of the domain must be 63 characters or less"
      };
    }
    if (label.startsWith("-") || label.endsWith("-")) {
      return {
        isValid: false,
        error: "Domain parts cannot start or end with hyphens"
      };
    }
  }

  if (!domainWithoutDot.includes(".")) {
    return {
      isValid: false,
      error: "Domain must contain at least one dot (e.g., example.com.)"
    };
  }

  const tld = labels[labels.length - 1];
  if (!/^[a-zA-Z]{2,}$/.test(tld)) {
    return {
      isValid: false,
      error:
        "Top-level domain must contain only letters and be at least 2 characters"
    };
  }

  return { isValid: true };
}

async function CreatePrefetch(domain: string, refresh: number, qtype: number) {
  const [code, response] = await PostRequest("prefetch", {
    domain,
    refresh,
    qtype
  });
  if (code === 200) {
    toast.success(`${domain} has been added to prefetch list!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

async function DeletePrefetch(domain: string) {
  const [code, response] = await DeleteRequest(
    `prefetch?domain=${domain}`,
    null
  );
  if (code === 200) {
    toast.success(`${domain} has been removed from prefetch list!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

export function Prefetch() {
  const [prefetches, setPrefetches] = useState<PrefetchEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [domainName, setDomainName] = useState("");
  const [refresh, setrefresh] = useState(0);
  const [qtype, setQType] = useState("1");
  const [searchTerm, setSearchTerm] = useState("");
  const [domainError, setDomainError] = useState<string>("");

  const fetchPrefetches = async () => {
    setLoading(true);
    const [code, response] = await GetRequest("prefetch");
    if (code !== 200) {
      toast.error("Unable to fetch DNS prefetch entries");
      setLoading(false);
      return;
    }

    setPrefetches(response || []);
    setLoading(false);
  };

  useEffect(() => {
    fetchPrefetches();
  }, []);

  useEffect(() => {
    if (domainName) {
      const validation = validateFQDN(domainName);
      setDomainError(validation.error || "");
    } else {
      setDomainError("");
    }
  }, [domainName]);

  const handleSave = async () => {
    const validation = validateFQDN(domainName);

    if (!validation.isValid) {
      toast.error(validation.error);
      setDomainError(validation.error || "");
      return;
    }

    setSubmitting(true);
    const success = await CreatePrefetch(domainName, refresh, parseInt(qtype));
    if (success) {
      await fetchPrefetches();
      setDomainName("");
      setDomainError("");
    }
    setSubmitting(false);
  };

  const handleDelete = async (domain: string) => {
    const success = await DeletePrefetch(domain);
    if (success) {
      await fetchPrefetches();
    }
  };

  const formatRefresh = (seconds: number) => {
    if (seconds === 0) return "On TTL Expire";
    if (seconds < 60) return `${seconds} seconds`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)} minutes`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours`;
    return `${Math.floor(seconds / 86400)} days`;
  };

  const filteredPrefetches = searchTerm
    ? prefetches.filter((prefetch) =>
        prefetch.domain.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : prefetches;

  const isFormValid = domainName && !domainError;

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            DNS Prefetch Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Pre-resolve domain names to improve the response time
          </p>
        </div>
        <div className="flex items-center gap-2">
          <DatabaseIcon className="h-3 w-3" />
          {prefetches.length} {prefetches.length === 1 ? "Entry" : "Entries"}
        </div>
      </div>

      <Card className="shadow-md">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2">
            <PlusIcon className="h-5 w-5 text-green-500" />
            Add DNS Prefetch
          </CardTitle>
          <CardDescription>
            DNS prefetching preemptively resolves domain names to IP addresses
            before they're needed. This can reduce page load times by
            eliminating DNS resolution delays when users navigate to prefetched
            domains.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="domain" className="font-medium">
                  Domain name
                </Label>
                <div className="relative">
                  <GlobeIcon className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                  <Input
                    id="domain"
                    placeholder="example.com."
                    className={`pl-9 ${
                      domainError ? "border-destructive" : ""
                    }`}
                    value={domainName}
                    onChange={(e) => setDomainName(e.target.value)}
                  />
                </div>
                {domainError && (
                  <span className="text-xs text-red-500 font-medium">
                    {domainError}
                    <br />
                  </span>
                )}
                <span className="text-xs text-muted-foreground">
                  Enter the domain you want to pre-resolve.
                </span>
                <span className="text-xs text-muted-foreground font-bold">
                  <br />
                  Note:{" "}
                </span>
                <span className="text-xs text-muted-foreground">
                  A dot must be added at the end in order for the domain to be
                  fully qualified (FQDN)
                </span>
              </div>
              <div className="space-y-2">
                <Label htmlFor="refresh" className="font-medium">
                  Refresh Interval
                </Label>
                <Select
                  value={refresh.toString()}
                  onValueChange={(value) => setrefresh(parseInt(value))}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select refresh interval" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">On TTL Expire</SelectItem>
                  </SelectContent>
                </Select>
                <div>
                  <span className="text-xs text-muted-foreground">
                    How often DNS records should be refreshed in the cache
                    <br />
                    'On TTL Expire' will re-fetch once the domain TTL expires
                  </span>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="qtype" className="font-medium">
                  Query Type
                </Label>
                <Select
                  value={qtype}
                  onValueChange={(value) => setQType(value)}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select query type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">A (IPv4 address)</SelectItem>
                    <SelectItem value="28">AAAA (IPv6 address)</SelectItem>
                    <SelectItem value="5">CNAME (Canonical name)</SelectItem>
                    <SelectItem value="12">PTR (Pointer record)</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  The type of DNS record to prefetch
                </p>
              </div>
            </div>
          </div>
        </CardContent>
        <div className="flex justify-end p-4">
          <Button
            variant="default"
            className="bg-green-600 hover:bg-green-700 text-white"
            onClick={handleSave}
            disabled={submitting || !isFormValid}
          >
            {submitting ? (
              <>
                <SpinnerIcon className="h-4 w-4 mr-2 animate-spin" />
                Adding...
              </>
            ) : (
              "Add Prefetch"
            )}
          </Button>
        </div>
      </Card>

      <Card className="shadow-md">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <ClockIcon className="h-5 w-5 text-blue-500" />
              Active Prefetch Domains
            </CardTitle>
            <div className="w-64">
              <Input
                placeholder="Search domains..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="text-sm"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-4">
          {loading ? (
            <div className="p-6 space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-center justify-between">
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-48" />
                    <Skeleton className="h-4 w-24" />
                  </div>
                  <Skeleton className="h-8 w-8 rounded-full" />
                </div>
              ))}
            </div>
          ) : filteredPrefetches.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Domain</TableHead>
                  <TableHead>Refresh Interval</TableHead>
                  <TableHead>Query Type</TableHead>
                  <TableHead className="text-right">Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPrefetches.map((prefetch) => (
                  <TableRow
                    key={prefetch.domain}
                    className="hover:bg-accent text-sm"
                  >
                    <TableCell className="font-medium">
                      {prefetch.domain}
                    </TableCell>
                    <TableCell className="font-mono">
                      {formatRefresh(prefetch.refresh)}
                    </TableCell>
                    <TableCell className="">
                      {qtypeExpanded(prefetch.qtype)}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-red-500 hover:text-red-700 hover:font-bold cursor-pointer"
                          onClick={() => handleDelete(prefetch.domain)}
                        >
                          <TrashIcon className="h-4 w-4" />
                          <span className="sr-only">Delete</span>
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <ClockIcon className="h-12 w-12 mb-4" />
              <h3 className="text-lg font-medium">No prefetch domains found</h3>
              <p className="text-muted-foreground mt-1">
                {searchTerm
                  ? "No matching entries for your search"
                  : "Add a domain to prefetch to get started"}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
