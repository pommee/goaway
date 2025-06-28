import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { PostRequest } from "@/util";
import { ArrowsClockwiseIcon } from "@phosphor-icons/react";
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

  const handleSave = () => {
    const domains = textareaValue
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line !== "");

    SendDomains(domains);
    setTextareaValue("");
    setDialogOpen(false);
  };

  return (
    <div className="mb-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogTrigger asChild>
          <Button variant="outline" onClick={() => setDialogOpen(true)}>
            <ArrowsClockwiseIcon className="mr-2" size={20} />
            Update custom
          </Button>
        </DialogTrigger>
        <DialogContent className="md:w-2/3">
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
          <Button onClick={handleSave}>Save</Button>
        </DialogContent>
      </Dialog>
    </div>
  );
}
