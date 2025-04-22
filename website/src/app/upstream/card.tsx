import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
  CardDescription
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { UpstreamEntry } from "@/pages/upstream";
import { PutRequest, DeleteRequest } from "@/util";
import { Cloud, Star } from "@phosphor-icons/react";
import { useState, useEffect } from "react";

export function UpstreamCard(upstream: UpstreamEntry) {
  const currentUpstream = upstream;
  const [isPreferred, setIsPreferred] = useState(upstream.preferred);
  const [deleteState, setDeleteState] = useState<"initial" | "confirm">(
    "initial"
  );

  useEffect(() => {
    let timeoutId: NodeJS.Timeout;

    if (deleteState === "confirm") {
      timeoutId = setTimeout(() => {
        setDeleteState("initial");
      }, 3000);
    }

    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [deleteState]);

  async function setPreferred(upstream: string) {
    try {
      const [status, response] = await PutRequest("preferredUpstream", {
        upstream: upstream
      });

      if (status === 200) {
        toast.info(response.message);
        setIsPreferred(true);
      } else {
        toast.warning(response.message);
      }
    } catch (error) {
      toast.error("Failed to set preferred upstream");
    }
  }

  async function handleDelete() {
    if (deleteState === "initial") {
      setDeleteState("confirm");
      return;
    }

    try {
      const [status, response] = await DeleteRequest(
        `upstream?upstream=${currentUpstream.upstream}`
      );

      if (status === 200) {
        toast.success(response.message);
      } else {
        toast.warning(response.message || "Failed to delete upstream");
      }
    } catch (error) {
      toast.error("Failed to delete upstream");
    } finally {
      setDeleteState("initial");
    }
  }

  return (
    <Card className="w-full max-w-sm bg-background/80 backdrop-blur-sm border-zinc-700/50 shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Cloud className="text-primary" size={24} />
            <span className="font-bold">{upstream.name}</span>
          </div>
        </CardTitle>
        <CardDescription>{upstream.upstream}</CardDescription>
      </CardHeader>

      <CardContent>
        <div className="flex items-center space-x-2 text-muted-foreground">
          <p>DNS Ping: </p>
          <p className="text-white">{upstream.dnsPing}</p>
        </div>
        <div className="flex items-center space-x-2 text-muted-foreground">
          <p>ICMP Ping:</p>
          <p className="text-white">{upstream.icmpPing}</p>
        </div>
      </CardContent>

      <CardFooter className="gap-2 grid lg:grid-cols-2">
        {isPreferred ? (
          <Button className="w-full text-white font-bold bg-green-700 hover:bg-green-700 cursor-default">
            <Star className="mr-2" size={16} />
            Preferred
          </Button>
        ) : (
          <Button
            className="w-full"
            onClick={() => setPreferred(upstream.upstream)}
            variant="secondary"
          >
            <Star className="mr-2" size={16} />
            Set Preferred
          </Button>
        )}
        <Button
          className={`${
            deleteState === "confirm"
              ? "bg-red-600 hover:bg-red-500"
              : "bg-red-800 hover:bg-red-600"
          } text-white relative overflow-hidden transition-all duration-300`}
          onClick={handleDelete}
        >
          <span
            className={`absolute inset-0 flex items-center justify-center transition-transform duration-300 ${
              deleteState === "confirm" ? "translate-y-0" : "translate-y-full"
            }`}
          >
            Confirm?
          </span>
          <span
            className={`transition-transform duration-300 ${
              deleteState === "confirm"
                ? "-translate-y-full opacity-0"
                : "translate-y-0 opacity-100"
            }`}
          >
            Delete
          </span>
        </Button>
      </CardFooter>
    </Card>
  );
}
