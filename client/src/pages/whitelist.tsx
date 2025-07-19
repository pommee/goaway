import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import {
  ClockIcon,
  DatabaseIcon,
  GlobeIcon,
  PlusIcon,
  ShieldCheckIcon,
  SpinnerIcon,
  TrashIcon
} from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

async function CreateWhitelistedDomain(domain: string) {
  const [code, response] = await PostRequest("whitelist", {
    domain: domain
  });
  if (code === 200) {
    toast.success(`${domain} is now whitelisted!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

async function DeleteWhitelistedDomain(domain: string) {
  const [code, response] = await DeleteRequest(
    `whitelist?domain=${domain}`,
    null
  );
  if (code === 200) {
    toast.success(`${domain} is no longer whitelisted!`);
    return true;
  } else {
    toast.error(response.error);
    return false;
  }
}

export function Whitelist() {
  const [whitelistedDomains, setwhitelistedDomains] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [domainName, setDomainName] = useState("");
  const [searchTerm, setSearchTerm] = useState("");

  const fetchWhitelistedDomains = async () => {
    setLoading(true);
    const [code, response] = await GetRequest("whitelist");
    if (code !== 200) {
      toast.error("Unable to fetch whitelisted domains");
      setLoading(false);
      return;
    }

    setwhitelistedDomains(response || []);
    setLoading(false);
  };

  useEffect(() => {
    fetchWhitelistedDomains();
  }, []);

  const handleSave = async () => {
    if (!domainName) {
      toast.warning("Domain is required");
      return;
    }

    setSubmitting(true);
    const success = await CreateWhitelistedDomain(domainName);
    if (success) {
      await fetchWhitelistedDomains();
      setDomainName("");
    }
    setSubmitting(false);
  };

  const handleDelete = async (domain: string) => {
    const success = await DeleteWhitelistedDomain(domain);
    if (success) {
      await fetchWhitelistedDomains();
    }
  };

  const filteredDomains = searchTerm
    ? whitelistedDomains.filter((domain) =>
        domain.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : whitelistedDomains;

  return (
    <div className="flex justify-center items-center">
      <div className="space-y-8 xl:w-2/3">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-4xl font-bold">Domain whitelist</h1>
            <p className="text-muted-foreground text-sm">
              Whitelisted domains will surpass blacklisted ones. Can be useful
              if a list were to block <strong>example.com</strong>, but you want
              it to always resolve.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <DatabaseIcon className="h-3 w-3" />
            {whitelistedDomains.length}{" "}
            {whitelistedDomains.length === 1 ? "Entry" : "Entries"}
          </div>
        </div>

        <Card>
          <CardHeader className="pb-4 border-b-2">
            <CardTitle className="flex items-center gap-2">
              <PlusIcon className="h-5 w-5 text-green-500" />
              New whitelisted domain
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid gap-4 md:grid-cols-4">
                <div className="md:col-span-3 space-y-2">
                  <Label htmlFor="domain" className="font-medium">
                    Domain name
                  </Label>
                  <div className="relative">
                    <GlobeIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="domain"
                      placeholder="example.com"
                      className="pl-9"
                      value={domainName}
                      onChange={(e) => setDomainName(e.target.value)}
                    />
                  </div>
                  <span className="text-xs text-muted-foreground">
                    Enter the domain you want to whitelist.
                  </span>
                </div>
                <div className="flex items-end mb-8">
                  <Button
                    variant="default"
                    className="cursor-pointer w-full bg-green-600 hover:bg-green-700"
                    onClick={handleSave}
                    disabled={submitting || !domainName}
                  >
                    {submitting ? (
                      <>
                        <SpinnerIcon className="h-4 w-4 mr-2 animate-spin" />
                        Adding...
                      </>
                    ) : (
                      "Add"
                    )}
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-4 border-b-2">
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <ShieldCheckIcon className="h-5 w-5 text-blue-500" />
                Whitelisted domains
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
          <CardContent>
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
            ) : filteredDomains.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow className="hover:bg-transparent">
                    <TableHead>Domain</TableHead>
                    <TableHead className="text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredDomains.map((domain) => (
                    <TableRow key={domain} className="hover:bg-accent">
                      <TableCell className="font-medium">{domain}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="cursor-pointer h-8 w-8 p-0 text-red-500 hover:text-red-700 hover:font-bold"
                            onClick={() => handleDelete(domain)}
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
              <div className="flex flex-col items-center justify-center py-6 text-center">
                <ClockIcon className="h-12 w-12 mb-4" />
                <h3 className="text-lg font-medium">
                  No whitelisted domains found
                </h3>
                <p className="text-muted-foreground mt-1">
                  {searchTerm
                    ? "No matching entries for your search"
                    : "Add a whitelisted domain to get started"}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
