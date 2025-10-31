import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { formatBytes, formatDate } from "./helpers";

export const ImportModal = ({
  open,
  onClose,
  onConfirm,
  filename
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  filename?: string;
}) => {
  const [fileDetails, setFileDetails] = useState<{
    name: string;
    size: number;
    lastModified: number;
  } | null>(null);

  useEffect(() => {
    if (filename) {
      const input = document.querySelector(
        "input[type='file']"
      ) as HTMLInputElement;
      const file = input?.files?.[0];
      if (file) {
        setTimeout(() => {
          setFileDetails({
            name: file.name,
            size: file.size,
            lastModified: file.lastModified
          });
        }, 0);
      }
    }
  }, [filename]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>Confirm Import</DialogTitle>
          <DialogDescription>
            <p>
              Replace the current database with <strong>{filename}</strong>?
            </p>
            <p className="mt-2">
              A backup of your current database will be created.
            </p>
            {fileDetails && (
              <div className="mt-4 p-2 rounded text-sm">
                <p>
                  <strong>File Details:</strong>
                </p>
                <ul className="mt-1 list-disc ml-4 space-y-1">
                  <li>
                    <strong>Name:</strong> {fileDetails.name}
                  </li>
                  <li>
                    <strong>Size:</strong> {formatBytes(fileDetails.size)}
                  </li>
                  <li>
                    <strong>Last Modified:</strong>{" "}
                    {formatDate(fileDetails.lastModified)}
                  </li>
                </ul>
              </div>
            )}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            className="hover:font-bold transition-all duration-200 bg-destructive/20"
          >
            Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
