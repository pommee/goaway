import { GetRequest } from "@/util";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export default function BlockingTimer() {
  const [timeLeft, setTimeLeft] = useState(0);

  useEffect(() => {
    async function fetchNotifications() {
      try {
        const [code, response] = await GetRequest("pause");
        if (code !== 200) {
          toast.warning("Unable to fetch blocking status", {
            id: "fetch-notifications-error"
          });
          return;
        }

        setTimeLeft(response.timeLeft || 0);
      } catch {
        toast.error("Error while fetching notifications");
      }
    }

    fetchNotifications();

    const intervalId = setInterval(() => {
      fetchNotifications();
    }, 1000);

    return () => clearInterval(intervalId);
  }, []);

  return (
    <div className="w-max">
      {timeLeft === 0 ? (
        <div className="text-xs font-medium text-green-500/80">
          Blocking active
        </div>
      ) : (
        <div className="text-xs font-medium text-red-500/80">
          Blocking paused:{" "}
          {Math.floor(timeLeft / 60)
            .toString()
            .padStart(2, "0")}
          :{(timeLeft % 60).toString().padStart(2, "0")}
        </div>
      )}
    </div>
  );
}
