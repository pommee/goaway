import { toast } from "sonner";
import { SpinnerIcon, UploadIcon, DownloadIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { formatBytes } from "./helpers";
import { getApiBaseUrl } from "@/util";
import { SettingRow } from "./SettingsRow";
import { SetModalsType } from "./types";

export const DatabaseSection = ({
  loading,
  setLoading,
  fileInput,
  setFile,
  setModals
}: {
  loading: { main: boolean; import: boolean; export: boolean };
  setLoading: React.Dispatch<
    React.SetStateAction<{ main: boolean; import: boolean; export: boolean }>
  >;
  fileInput: React.RefObject<HTMLInputElement>;
  setFile: (file: File | null) => void;
  setModals: React.Dispatch<React.SetStateAction<SetModalsType>>;
}) => {
  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const uploadedFile = e.target.files?.[0];
    if (uploadedFile?.name.endsWith(".db")) {
      setFile(uploadedFile);
      setModals((prev: SetModalsType) => ({ ...prev, importConfirm: true }));
    } else {
      toast.error("Please select a .db file");
    }
  };

  const exportDb = async () => {
    setLoading((prev) => ({ ...prev, export: true }));
    const toastId = toast.loading("Starting export...", {
      description: "Preparing database for export",
      duration: Infinity
    });

    try {
      const response = await fetch(`${getApiBaseUrl()}/api/exportDatabase`);
      if (!response.ok)
        throw new Error(`Export failed: ${response.statusText}`);
      if (!response.body) throw new Error("ReadableStream not supported");

      const total = parseInt(response.headers.get("Content-Length") || "0", 10);
      const reader = response.body.getReader();
      const chunks = [];
      let received = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        chunks.push(value);
        received += value.length;

        const progressText =
          total > 0
            ? `${Math.round((received / total) * 100)}% (${formatBytes(
                received
              )} / ${formatBytes(total)})`
            : `Downloaded ${formatBytes(received)}`;

        toast.loading("Downloading database...", {
          id: toastId,
          description: progressText,
          duration: Infinity
        });
      }

      const blob = new Blob(chunks, { type: "application/octet-stream" });
      const url = URL.createObjectURL(blob);
      const a = Object.assign(document.createElement("a"), {
        href: url,
        download: "database.db"
      });
      document.body.appendChild(a).click();
      a.remove();
      URL.revokeObjectURL(url);

      toast.success("Database exported successfully!", {
        id: toastId,
        description: `Downloaded ${formatBytes(received)}`,
        duration: 4000
      });
    } catch (error) {
      toast.error("Export failed", {
        id: toastId,
        description: error instanceof Error ? error.message : "An error occurred during export",
        duration: 5000
      });
    } finally {
      setLoading((prev) => ({ ...prev, export: false }));
    }
  };

  return (
    <>
      <SettingRow
        title="Export database"
        description="Download current database file"
        action={
          <Button
            variant="outline"
            onClick={exportDb}
            disabled={loading.import || loading.export}
          >
            {loading.export ? (
              <>
                <SpinnerIcon className="animate-spin mr-2" /> Exporting...
              </>
            ) : (
              <>
                <UploadIcon className="mr-2" /> Export
              </>
            )}
          </Button>
        }
      />
      <SettingRow
        title="Import database"
        description="Replace current database (backup created)"
        action={
          <>
            <input
              ref={fileInput}
              type="file"
              accept=".db"
              onChange={handleFileUpload}
              className="hidden"
            />
            <Button
              variant="outline"
              onClick={() => fileInput.current?.click()}
              disabled={loading.import || loading.export}
            >
              {loading.import ? (
                <>
                  <SpinnerIcon className="animate-spin mr-2" /> Importing...
                </>
              ) : (
                <>
                  <DownloadIcon className="mr-2" /> Import
                </>
              )}
            </Button>
          </>
        }
      />
    </>
  );
};
