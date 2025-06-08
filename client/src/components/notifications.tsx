import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import { DeleteRequest, GetRequest } from "@/util";
import { Bell, Info, Warning } from "@phosphor-icons/react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

type NotificationsResponse = {
  id: number;
  severity: string;
  category: string;
  text: string;
  read: boolean;
  createdAt: string;
};

export default function Notifications() {
  const [notifications, setNotifications] = useState<NotificationsResponse[]>(
    []
  );
  const [open, setOpen] = useState(false);
  const prevUnreadCountRef = useRef(0);

  const unreadCount = notifications.filter(
    (notification) => !notification.read
  ).length;

  const shouldAnimate = unreadCount > prevUnreadCountRef.current;

  useEffect(() => {
    prevUnreadCountRef.current = unreadCount;
  }, [unreadCount]);

  useEffect(() => {
    async function fetchNotifications() {
      try {
        const [code, response] = await GetRequest("notifications");
        if (code !== 200) {
          toast.warning("Unable to fetch notifications");
          return;
        }

        setNotifications(response.notifications);
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

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case "error":
        return <Warning className="h-5 w-5 text-red-500" />;
      case "warning":
        return <Warning className="h-5 w-5 text-amber-500" />;
      case "info":
      default:
        return <Info className="h-5 w-5 text-blue-500" />;
    }
  };

  const getSeverityColorClass = (severity: string) => {
    switch (severity) {
      case "error":
        return "border-l-4 border-red-500 bg-red-500/10";
      case "warning":
        return "border-l-4 border-amber-500 bg-amber-500/10";
      case "info":
      default:
        return "border-l-4 border-blue-500 bg-blue-500/10";
    }
  };

  const getTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.round(diffMs / 1000);
    const diffMins = Math.round(diffSecs / 60);
    const diffHours = Math.round(diffMins / 60);
    const diffDays = Math.round(diffHours / 24);

    if (diffSecs < 60) return `${diffSecs} seconds ago`;
    if (diffMins < 60) return `${diffMins} minutes ago`;
    if (diffHours < 24) return `${diffHours} hours ago`;
    return `${diffDays} days ago`;
  };

  const handleMarkAllAsRead = async (e) => {
    e.stopPropagation();
    const updatedNotifications = notifications.map((notification) => ({
      ...notification,
      read: true
    }));
    setNotifications(updatedNotifications);
    try {
      const notificationIds = updatedNotifications.map(
        (notification) => notification.id
      );
      const [status, response] = await DeleteRequest("notification", {
        notificationIds
      });

      if (status !== 200) {
        toast.success(response.error);
      }
    } catch {
      toast.error("Failed to mark notifications as read");
    }
  };

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild className="cursor-pointer">
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <div className="absolute -top-1 -right-1 overflow-hidden">
              <div
                className={`flex items-center justify-center min-w-5 h-5 rounded-full bg-red-500 text-white text-xs font-medium ${
                  shouldAnimate ? "animate-pulse" : ""
                }`}
              >
                {unreadCount}
              </div>
            </div>
          )}
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-96 p-0 bg-stone-950 border border-stone-800 shadow-xl rounded-lg"
      >
        <DropdownMenuLabel className="flex justify-between items-center p-4 border-b border-stone-800">
          <span className="font-semibold text-lg">Notifications</span>
          {notifications.some((n) => !n.read) && (
            <Button
              variant="outline"
              size="sm"
              className="h-8 text-xs hover:bg-stone-800"
              onClick={handleMarkAllAsRead}
            >
              Mark all as read
            </Button>
          )}
        </DropdownMenuLabel>

        <ScrollArea className="h-96">
          {notifications.length > 0 ? (
            <div className="divide-y divide-stone-800">
              {notifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`p-4 ${getSeverityColorClass(
                    notification.severity
                  )} ${
                    notification.read ? "opacity-70" : ""
                  } transition-all duration-200`}
                >
                  <div className="flex items-start gap-3">
                    <div className="mt-1">
                      {getSeverityIcon(notification.severity)}
                    </div>
                    <div className="flex-1 space-y-1">
                      <p
                        className={`text-sm ${
                          notification.read ? "" : "font-medium"
                        }`}
                      >
                        {notification.text}
                      </p>
                      <div className="flex justify-between items-center">
                        <p className="text-xs text-stone-400">
                          {getTimeAgo(notification.createdAt)}
                        </p>
                        <span className="text-xs px-2 py-1 rounded-full bg-stone-800 text-stone-300">
                          {notification.category}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-32 text-stone-400 p-4">
              <Bell className="h-6 w-6 mb-2 opacity-50" />
              <p>No notifications</p>
            </div>
          )}
        </ScrollArea>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
