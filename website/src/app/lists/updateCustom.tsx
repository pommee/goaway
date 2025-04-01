import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { RefreshCcw } from "lucide-react";
import { DialogDescription } from "@radix-ui/react-dialog";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { PostRequest } from "@/util";

async function SendDomains(domains: string[]) {
  try {
    const response = await PostRequest("/api/custom", domains);
    if (response.ok) {
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

  const handleSave = () => {
    const domains = textareaValue
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line !== "");

    SendDomains(domains);
    setDialogOpen(false);
  };

  return (
    <div className="mb-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className="bg-zinc-800 border-none hover:bg-zinc-700 text-white"
            onClick={() => setDialogOpen(true)}
          >
            <RefreshCcw className="mr-2" size={20} />
            Update custom
          </Button>
        </DialogTrigger>
        <DialogContent className="bg-zinc-900 text-white border-zinc-800 w-1/3 max-w-none">
          <DialogHeader>
            <DialogTitle>Update custom</DialogTitle>
          </DialogHeader>
          <DialogDescription className="text-base leading-relaxed">
            You can maintain custom addresses using a custom list. Simply add
            the domain and it will be blocked.
          </DialogDescription>
          <Textarea
            value={textareaValue}
            onChange={(e) => setTextareaValue(e.target.value)}
            placeholder="ads-are-boring.com&#10;remove-this-domain.gov&#10;whatever.you.want.com&#10;..."
          />
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
