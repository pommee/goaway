"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { NoContent } from "@/shared";
import { GetRequest } from "@/util";
import { ArticleIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

type AuditEntry = {
  id: number;
  topic: string;
  message: string;
  createdAt: string;
};

export default function Audit() {
  const [audits, setAudits] = useState<AuditEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchAudits = async () => {
      try {
        const [status, response] = await GetRequest("audit");
        if (status === 200) setAudits(response);
      } catch (error) {
        toast.warning("Failed to fetch audits", {
          description: response.error
        });
        console.error("Failed to fetch audits:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchAudits();
    const interval = setInterval(fetchAudits, 5000);
    return () => clearInterval(interval);
  }, []);

  const formatDate = (dateString: string) => {
    if (!dateString) return "Never";

    try {
      return new Date(dateString).toLocaleString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false
      });
    } catch {
      return dateString;
    }
  };

  return (
    <Card className="max-h-64 overflow-x-hidden gap-0 py-0 pt-2 xl:w-1/2  mt-5 pb-2">
      <CardHeader className="pl-4 mb-2">
        <CardTitle className="flex items-center gap-2">
          <ArticleIcon size={18} />
          Audit Log
        </CardTitle>
      </CardHeader>

      <CardContent className="p-0">
        {isLoading ? (
          <div className="p-4 text-center text-muted-foreground">
            Loading...
          </div>
        ) : audits.length > 0 ? (
          <div className="space-y-2">
            {audits.map((audit) => (
              <div
                key={audit.id}
                className="px-4 py-2 hover:bg-accent/50 transition-colors border-l border-orange-500/20 mx-2"
              >
                <div className="flex justify-between">
                  <span className="text-xs text-muted-foreground bg-orange-500/20 px-2 py-1 rounded">
                    {audit.topic}
                  </span>
                  <time className="text-xs text-muted-foreground">
                    {formatDate(audit.createdAt)}
                  </time>
                </div>
                <p className="mt-1 text-sm">{audit.message}</p>
              </div>
            ))}
          </div>
        ) : (
          <NoContent text={"No audits created"} />
        )}
      </CardContent>
    </Card>
  );
}
