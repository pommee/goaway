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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import { Database, Globe, Plus, Spinner, Trash } from "@phosphor-icons/react";
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
    `resolution?domain=${domain}&ip=${ip}`
  );
  if (code === 200) {
    toast.success(`${domain} has been deleted!`);
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
    const [code, response] = await GetRequest(`resolutions`);
    if (code !== 200) {
      toast.error(`Unable to fetch resolutions`);
      setLoading(false);
      return;
    }

    const listArray: ListEntry[] = Object.entries(
      response.resolutions || {}
    ).map(([, details]) => ({
      domain: details.domain,
      ip: details.ip
    }));

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
          <p className="text-gray-500 mt-1">
            Map custom domains to specific IP addresses
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Database className="h-3 w-3" />
          {resolutions.length} {resolutions.length === 1 ? "Entry" : "Entries"}
        </div>
      </div>

      <Card className="shadow-md border-green-100">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2">
            <Plus className="h-5 w-5 text-green-500" />
            Add New Resolution
          </CardTitle>
          <CardDescription>
            Create a custom domain-to-IP mapping for your network
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="domain" className="font-medium">
                Domain name
              </Label>
              <div className="relative">
                <Globe className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                <Input
                  id="domain"
                  placeholder="example.local"
                  className="pl-9"
                  value={domainName}
                  onChange={(e) => setDomainName(e.target.value)}
                />
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
            </div>
          </div>
        </CardContent>
        <div className="flex justify-end items-end">
          <Button
            variant="default"
            className="mr-5 bg-green-600 hover:bg-green-700 text-white"
            onClick={handleSave}
            disabled={submitting || !domainName || !ip}
          >
            {submitting ? (
              <>
                <Spinner className="h-4 w-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : (
              "Save Resolution"
            )}
          </Button>
        </div>
      </Card>

      <Card className="shadow-md">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5 text-blue-500" />
              Current Resolutions
            </CardTitle>
            <div className="w-64">
              <Input
                placeholder="Search domains or IPs..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="text-sm"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
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
          ) : filteredResolutions.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Domain</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead className="text-right">Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredResolutions.map((resolution) => (
                  <TableRow
                    key={resolution.domain}
                    className="hover:bg-stone-800"
                  >
                    <TableCell className="font-medium text-white">
                      {resolution.domain}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {resolution.ip}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-red-500 hover:text-red-700 hover:bg-stone-700"
                        onClick={() =>
                          handleDelete(resolution.domain, resolution.ip)
                        }
                      >
                        <Trash className="h-4 w-4" />
                        <span className="sr-only">Delete</span>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Database className="h-12 w-12 text-gray-300 mb-4" />
              <h3 className="text-lg font-medium text-gray-400">
                No resolutions found
              </h3>
              <p className="text-gray-500 mt-1">
                {searchTerm
                  ? "No matching entries for your search"
                  : "Add a new resolution to get started"}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
