import { useEffect, useState } from "react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Clock, Eye, User } from "lucide-react";
import { CardDetails } from "./details";
import { ClientEntry } from "@/pages/clients";
import { Button } from "@/components/ui/button";

function timeAgo(timestamp: string) {
  const now = new Date();
  const past = new Date(timestamp);
  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000);

  const seconds = diffInSeconds % 60;
  const minutes = Math.floor((diffInSeconds / 60) % 60);
  const hours = Math.floor(diffInSeconds / 3600);

  return `${hours}h ${minutes}m ${seconds}s ago`;
}

export function ClientCard(clientEntry: ClientEntry) {
  const [lastSeenText, setLastSeenText] = useState(() =>
    timeAgo(clientEntry.lastSeen)
  );
  const [showDetails, setShowDetails] = useState(false);

  useEffect(() => {
    setLastSeenText(timeAgo(clientEntry.lastSeen));
  }, [clientEntry.lastSeen]);

  return (
    <>
      <Card className="w-full max-w-sm bg-background/80 backdrop-blur-sm border-zinc-700/50 shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <User className="text-primary" size={24} />
              <span className="text-xl font-bold">{clientEntry.name}</span>
            </div>
          </CardTitle>
        </CardHeader>

        <Separator className="mb-4 bg-zinc-700/30" />

        <CardContent className="space-y-3">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-2 text-muted-foreground">
              <span className="text-sm font-medium">IP Address:</span>
              <code className="bg-secondary/50 px-2 py-1 rounded-md text-xs">
                {clientEntry.ip}
              </code>
            </div>

            <div className="flex items-center text-muted-foreground space-x-1">
              <Clock className="text-primary" size={14} />
              <span className="text-sm">{lastSeenText}</span>
            </div>
          </div>
        </CardContent>

        <CardFooter>
          <Button
            onClick={() => setShowDetails(true)}
            className="w-full group"
            variant="outline"
          >
            <Eye className="mr-2 group-hover:animate-pulse" size={16} />
            View Details
          </Button>
        </CardFooter>
      </Card>

      {showDetails && (
        <CardDetails
          ip={clientEntry.ip}
          lastSeen={clientEntry.lastSeen}
          mac={clientEntry.mac}
          name={clientEntry.name}
          vendor={clientEntry.vendor}
          onClose={() => setShowDetails(false)}
        />
      )}
    </>
  );
}
