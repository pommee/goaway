import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UpstreamEntry } from "@/pages/upstream";
import { PostRequest } from "@/util";
import { PlusIcon } from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { toast } from "sonner";

type AddUpstreamProps = {
  onAdd: (entry: UpstreamEntry) => void;
};

export function AddUpstream({ onAdd }: AddUpstreamProps) {
  const [newUpstreamIP, setNewUpstreamIP] = useState("");
  const [open, setOpen] = useState(false);

  const handleSave = async () => {
    if (!/^[\d\\.]+:\d+$/.test(newUpstreamIP.trim())) {
      toast.error(
        "Please enter the upstream in IP:PORT format, e.g. 1.1.1.1:53"
      );
      return;
    }
    const [code, response] = await PostRequest("upstream", {
      upstream: newUpstreamIP
    });
    if (code === 200) {
      toast.info(response.message);
      setOpen(false);
      onAdd({
        dnsPing: "reload to ping",
        icmpPing: "reload to ping",
        name: newUpstreamIP,
        preferred: false,
        upstream: newUpstreamIP
      });
      setNewUpstreamIP("");
    }
  };

  return (
    <div className="mb-5">
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button variant="default">
            <PlusIcon className="mr-2" size={20} />
            Add upstream
          </Button>
        </DialogTrigger>
        <DialogContent className="lg:w-1/3">
          <DialogHeader>
            <DialogTitle>New Upstream</DialogTitle>
          </DialogHeader>
          <DialogDescription className="text-base leading-relaxed">
            <p>
              A new upstream can be created by specifying the DNS server IP.
            </p>
            <span>
              Default is{" "}
              <span className="bg-accent p-0.5 pl-1 pr-1 rounded-sm">
                1.1.1.1 (Cloudflare)
              </span>
              and{" "}
              <span className="bg-accent p-0.5 pl-1 pr-1 rounded-sm">
                8.8.8.8 (Google)
              </span>
            </span>
          </DialogDescription>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="ip">DNS IP</Label>
              <Input
                id="ip"
                value={newUpstreamIP}
                placeholder="1.1.1.1:53"
                onChange={(e) => setNewUpstreamIP(e.target.value)}
                className="col-span-3"
              />
            </div>
            <span className="text-xs text-muted-foreground col-span-4 pl-2">
              Please enter the IP and port, e.g.{" "}
              <span className="font-mono">1.1.1.1:53</span>
            </span>
          </div>
          <Button
            variant="outline"
            className="bg-green-800 border-none hover:bg-green-600 text-white"
            onClick={handleSave}
          >
            Save
          </Button>
        </DialogContent>
      </Dialog>
    </div>
  );
}
