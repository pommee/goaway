import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { PostRequest } from "@/util";
import {
  ArrowsClockwiseIcon,
  ShieldIcon,
  InfoIcon
} from "@phosphor-icons/react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { toast } from "sonner";

async function SendDomains(domains: string[]) {
  try {
    const [status] = await PostRequest("custom", { domains: domains });
    if (status === 200) {
      toast.success("Domains updated successfully!");
    } else {
      toast.error("Failed to update domains.");
    }
  } catch {
    toast.error("An error occurred while sending the request.");
  }
}

export function UpdateCustom() {
  const [textareaValue, setTextareaValue] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSave = async () => {
    const domains = textareaValue
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line !== "");

    if (domains.length === 0) {
      toast.error("Please enter at least one domain.");
      return;
    }

    setIsLoading(true);
    try {
      await SendDomains(domains);
      setTextareaValue("");
      setDialogOpen(false);
    } finally {
      setIsLoading(false);
    }
  };

  const domainCount = textareaValue
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "").length;

  return (
    <div className="mb-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className="shadow-sm hover:shadow-md transition-all duration-200 border-2 hover:border-blue-300"
            onClick={() => setDialogOpen(true)}
          >
            <ArrowsClockwiseIcon className="mr-2" size={18} />
            Update Custom
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-hidden">
          <DialogHeader className="space-y-4 pb-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-full">
                <ShieldIcon size={24} className="text-blue-600" />
              </div>
              <div>
                <DialogTitle className="text-xl font-semibold">
                  Custom Domain Blocking
                </DialogTitle>
                <DialogDescription className="text-sm text-muted-foreground mt-1">
                  Manage your custom blocklist
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>

          <div className="space-y-6">
            <div className="bg-blue-800/40 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <InfoIcon
                  size={20}
                  className="text-blue-400 mt-0.5 flex-shrink-0"
                />
                <div className="text-sm text-blue-400">
                  <p className="font-medium mb-1">How it works:</p>
                  <p>
                    Add domain names (one per line) to your custom blocklist.
                    These domains will be automatically blocked across your
                    network.
                  </p>
                </div>
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label
                  htmlFor="domains-textarea"
                  className="text-sm font-medium text-muted-foreground"
                >
                  Domain List
                </Label>
                {domainCount > 0 && (
                  <span className="text-xs text-muted-foreground bg-accent px-2 py-1 rounded-full">
                    {domainCount} domain{domainCount !== 1 ? "s" : ""}
                  </span>
                )}
              </div>

              <Textarea
                id="domains-textarea"
                className="min-h-[200px] max-h-[300px] resize-none font-mono"
                value={textareaValue}
                onChange={(e) => setTextareaValue(e.target.value)}
                placeholder="ads-are-boring.com&#10;remove-this-domain.gov&#10;unwanted-tracker.net&#10;spam-domain.org&#10;..."
              />
            </div>
          </div>

          <DialogFooter className="flex flex-col sm:flex-row gap-4">
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={isLoading}
              className="order-2 sm:order-1"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={isLoading || domainCount === 0}
              className="order-1 sm:order-2 bg-blue-600 hover:bg-blue-700 text-white shadow-sm hover:shadow-md transition-all duration-200"
            >
              {isLoading ? (
                <>
                  <ArrowsClockwiseIcon className="mr-2 h-4 w-4 animate-spin" />
                  Updating...
                </>
              ) : (
                <>
                  <ShieldIcon className="mr-2 h-4 w-4" />
                  Update Blocklist
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
