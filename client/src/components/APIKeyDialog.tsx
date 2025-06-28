import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { GetRequest, PostRequest } from "@/util";
import { CopyIcon, KeyIcon, TrashIcon } from "@phosphor-icons/react";
import { DialogDescription, DialogTitle } from "@radix-ui/react-dialog";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { ScrollArea } from "./ui/scroll-area";

interface APIKey {
  name: string;
  key: string;
  createdAt: string;
}

interface NewAPIKeyData {
  key: string;
  id: string;
}

interface APIKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function APIKeyDialog({ open, onOpenChange }: APIKeyDialogProps) {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [newKeyName, setNewKeyName] = useState("");
  const [newKey, setNewKey] = useState<NewAPIKeyData | null>(null);

  const fetchAPIKeys = async () => {
    setIsLoading(true);
    try {
      const [status, response] = await GetRequest("apiKey");
      if (status === 200 && response) {
        setApiKeys(response.keys);
      }
    } catch (error) {
      console.error("Failed to fetch API keys:", error);
      toast.error("Could not load API keys");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (open) {
      fetchAPIKeys();
      setNewKey(null);
    }
  }, [open]);

  const handleGenerateKey = async () => {
    if (!newKeyName.trim()) {
      toast.error("Please enter a name for the API key");
      return;
    }

    try {
      const [status, response] = await PostRequest("apiKey", {
        name: newKeyName
      });
      if (status === 200 && response) {
        setNewKey(response);
        fetchAPIKeys();
        setNewKeyName("");
        toast.success("API key generated successfully");
      }
    } catch (error) {
      console.error("Failed to generate API key:", error);
      toast.error("Could not generate API key");
    }
  };

  const handleDeleteKey = async (key: string) => {
    try {
      const [status, message] = await GetRequest(`deleteApiKey?key=${key}`);
      if (status === 200) {
        toast.success(message.message);
        await fetchAPIKeys();
      }
    } catch (error) {
      console.error("Failed to delete API key:", error);
      toast.error("Could not delete API key");
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "Never";

    try {
      return new Date(dateString).toLocaleString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false
      });
    } catch {
      return dateString;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:w-2/3">
        <DialogHeader>
          <DialogTitle className="flex">
            <KeyIcon className="mt-1 mr-2" />
            API keys
          </DialogTitle>
        </DialogHeader>
        <DialogDescription className="text-sm leading-relaxed">
          <span>
            <span>
              API keys provide a way to establish a long-lived authenticated
              session with the server. After generating an API key, include it
              in your requests by setting the
            </span>
            <span className="bg-accent p-0.5 rounded-sm"> api-key </span>
            <span>header to the key's value.</span>
          </span>
        </DialogDescription>
        <div className="space-y-6 py-4">
          <div className="space-y-4">
            <h3 className="text-sm font-medium">Generate New API Key</h3>
            <div className="flex space-x-2">
              <Input
                placeholder="API Key Name"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="flex-1"
              />
              <Button onClick={handleGenerateKey}>Generate</Button>
            </div>
          </div>

          {newKey && (
            <div className="p-4 border rounded-md space-y-2">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">New API Key</h4>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => copyToClipboard(newKey.key)}
                >
                  <CopyIcon size={16} />
                </Button>
              </div>
              <p className="text-xs break-all bg-accent p-2 rounded">
                {newKey.key}
              </p>
              <p className="text-xs text-yellow-400">
                Save this key now. It will never be shown again!
              </p>
            </div>
          )}

          <div className="space-y-4">
            <h3 className="text-sm font-medium">Existing API Keys</h3>
            {isLoading ? (
              <div className="text-center py-4">Loading API keys...</div>
            ) : apiKeys.length === 0 ? (
              <div className="text-center py-4 text-sm text-muted-foreground">
                No API keys found
              </div>
            ) : (
              <ScrollArea className="h-[200px]">
                <div className="space-y-2">
                  {apiKeys.map((key) => (
                    <div
                      key={key.key}
                      className="flex justify-between items-center p-3 border rounded-md"
                    >
                      <div className="space-y-1">
                        <p className="font-medium text-sm">{key.name}</p>
                        <p className="text-xs text-muted-foreground">
                          Created: {formatDate(key.createdAt)}
                        </p>
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDeleteKey(key.key)}
                        className="text-red-500 hover:text-red-700"
                      >
                        <TrashIcon size={16} />
                      </Button>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
