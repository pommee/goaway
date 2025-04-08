import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
  Table
} from "@/components/ui/table";
import { DeleteRequest, GetRequest, PostRequest } from "@/util";
import { useEffect, useState } from "react";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter
} from "@/components/ui/card";
import { Trash } from "@phosphor-icons/react";

export type ListEntry = {
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
  } else {
    toast.warning(response.error);
  }
}

async function DeleteResolution(domain: string, ip: string) {
  const [code, response] = await DeleteRequest(
    `resolution?domain=${domain}&ip=${ip}`
  );
  if (code === 200) {
    toast.success(`${domain} has been deleted!`);
  } else {
    toast.warning(response.error);
  }
}

export function Resolution() {
  const [resolutions, setResolutions] = useState<ListEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [domainName, setDomainName] = useState("");
  const [ip, setIP] = useState("");

  const fetchResolutions = async () => {
    setLoading(true);
    const [code, response] = await GetRequest(`resolutions`);
    if (code !== 200) {
      toast.warning(`Unable to fetch resolutions`);
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

    await CreateResolution(domainName, ip);
    await fetchResolutions();
    setDomainName("");
    setIP("");
  };

  const handleDelete = async (domain: string, ip: string) => {
    await DeleteResolution(domain, ip);
    await fetchResolutions();
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Add New Resolution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="domain" className="text-right font-medium">
                Domain name
              </Label>
              <Input
                id="domain"
                placeholder="SomeCustomDomain"
                className="col-span-3"
                value={domainName}
                onChange={(e) => setDomainName(e.target.value)}
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="ip" className="text-right font-medium">
                IP
              </Label>
              <Input
                id="ip"
                placeholder="151.25.122.21"
                className="col-span-3"
                value={ip}
                onChange={(e) => setIP(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex justify-end">
          <Button
            variant="default"
            className="bg-green-600 hover:bg-green-700 text-white"
            onClick={handleSave}
          >
            Save Resolution
          </Button>
        </CardFooter>
      </Card>

      <Card className="shadow-md">
        <CardHeader className="pb-2">
          <CardTitle>Current Resolutions</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-4">Loading resolutions...</div>
          ) : resolutions.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Domain</TableHead>
                  <TableHead>IP</TableHead>
                  <TableHead className="text-right">Action</TableHead>{" "}
                </TableRow>
              </TableHeader>
              <TableBody>
                {resolutions.map((resolution) => (
                  <TableRow key={resolution.domain}>
                    <TableCell className="font-medium">
                      {resolution.domain}
                    </TableCell>
                    <TableCell>{resolution.ip}</TableCell>
                    <TableCell className="text-right">
                      {" "}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0"
                        onClick={() =>
                          handleDelete(resolution.domain, resolution.ip)
                        }
                      >
                        <Trash className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-4 text-gray-500">
              No resolutions found
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
