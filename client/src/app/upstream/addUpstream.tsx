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
import { PostRequest } from "@/util";
import { PlusIcon } from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { toast } from "sonner";

export function AddUpstream() {
  const [newUpstreamIP, setNewUpstreamIP] = useState("");
  const [open, setOpen] = useState(false);

  const handleSave = async () => {
    const [code, response] = await PostRequest(`upstream`, {
      upstream: newUpstreamIP
    });
    if (code === 200) {
      toast.info(response.message);
      setOpen(false);
    } else {
      toast.error(response.message || "Failed to create upstream");
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
                1.1.1.1 (Google)
              </span>
              and{" "}
              <span className="bg-accent p-0.5 pl-1 pr-1 rounded-sm">
                8.8.8.8 (Cloudflare)
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
