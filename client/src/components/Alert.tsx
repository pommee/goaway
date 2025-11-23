import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import {
  DiscordLogoIcon,
  MailboxIcon,
  NotificationIcon,
  SlackLogoIcon,
  TelegramLogoIcon,
  TestTubeIcon
} from "@phosphor-icons/react";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Checkbox } from "@/components/ui/checkbox";
import { GetRequest, PostRequest } from "@/util";
import { toast } from "sonner";

interface AlertsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

type AlertDiscordSettings = {
  enabled: boolean;
  name: string;
  webhook: string;
};

const DEFAULT_SETTINGS: AlertDiscordSettings = {
  enabled: false,
  name: "GoAway",
  webhook: ""
};

async function saveAlerts(
  onOpenChange: (open: boolean) => void,
  discord: AlertDiscordSettings
) {
  try {
    const [code, response] = await PostRequest("alert", { discord: discord });

    if (code === 200) {
      onOpenChange(false);
    } else {
      toast.error("Failed to save alerts", {
        description: `${response.error}`
      });
    }
  } catch (error) {
    toast.error("Error saving alerts", {
      description: `${error}`
    });
  }
}

async function testAlert(discord: AlertDiscordSettings) {
  try {
    const [code, response] = await PostRequest("alert/test", {
      discord: discord
    });

    if (code !== 200) {
      toast.error("Failed to send test alert", {
        description: `${response.error}`
      });
    }
  } catch (error) {
    toast.error("Error sending test alert", {
      description: `${error}`
    });
  }
}

async function fetchAlertSettings(): Promise<AlertDiscordSettings> {
  try {
    const [code, response] = await GetRequest("alert");

    if (code === 200) {
      return (response as { discord: AlertDiscordSettings }).discord;
    } else {
      toast.error("Failed to fetch alert settings", {
        description: response && typeof response === 'object' && 'error' in response ? String((response as { error?: unknown }).error) : 'Unknown error'
      });
      return DEFAULT_SETTINGS;
    }
  } catch (error) {
    toast.error("Error fetching alert settings", {
      description: `${error}`
    });
    return DEFAULT_SETTINGS;
  }
}

export function AlertDialog({ open, onOpenChange }: AlertsDialogProps) {
  const [discordSettings, setDiscordSettings] =
    useState<AlertDiscordSettings>(DEFAULT_SETTINGS);

  useEffect(() => {
    if (open) {
      fetchAlertSettings().then((settings) => {
        setDiscordSettings(settings);
      });
    }
  }, [open]);

  const handleSave = () => {
    saveAlerts(onOpenChange, discordSettings);
  };

  const handleTestWebhook = () => {
    testAlert(discordSettings);
  };

  const updateDiscordSettings = (updates: Partial<AlertDiscordSettings>) => {
    setDiscordSettings((prev) => ({ ...prev, ...updates }));
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:w-2/3 lg:w-2/4 bg-transparent backdrop-blur-sm">
        <DialogHeader>
          <DialogTitle className="flex">
            <NotificationIcon className="mr-2" />
            Alerts
          </DialogTitle>
        </DialogHeader>
        <DialogDescription className="text-sm leading-relaxed">
          <span>
            Alerts is a way for you to stay informed about important updates and
            events. You can customize your alert preferences to receive alerts
            via email or webhooks. This helps you stay connected and ensures you
            never miss out on important information. <br />
            <br />
            <b>Discord is currently only supported</b>
          </span>
        </DialogDescription>

        <div className="flex w-full flex-col gap-6">
          <Tabs defaultValue="discord">
            <TabsList>
              <TabsTrigger value="discord">
                <DiscordLogoIcon />
                Discord
              </TabsTrigger>
              <TabsTrigger value="slack" disabled={true}>
                <SlackLogoIcon />
                Slack
              </TabsTrigger>
              <TabsTrigger value="telegram" disabled={true}>
                <TelegramLogoIcon />
                Telegram
              </TabsTrigger>
              <TabsTrigger value="email" disabled={true}>
                <MailboxIcon />
                Email
              </TabsTrigger>
            </TabsList>
            <TabsContent value="discord">
              <Card>
                <CardContent className="grid gap-6">
                  <div className="grid grid-cols-2 items-center">
                    <div className="flex">
                      <p className="text-sm text-muted-foreground font-bold">
                        Enabled
                      </p>
                      <p className="ml-1 text-red-500">*</p>
                    </div>
                    <Checkbox
                      id="discord-enabled"
                      checked={discordSettings.enabled}
                      onCheckedChange={(checked) =>
                        updateDiscordSettings({ enabled: checked as boolean })
                      }
                    />
                  </div>

                  <div className="grid grid-cols-2 items-start">
                    <div className="text-left">
                      <div className="flex">
                        <p className="text-sm text-muted-foreground font-bold">
                          Webhook URL
                        </p>
                        <p className="ml-1 text-red-500">*</p>
                      </div>
                      <p className="text-muted-foreground text-xs mt-1">
                        Create a{" "}
                        <a
                          className="text-white"
                          target="_blank"
                          href="https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks"
                        >
                          webhook integration{" "}
                        </a>
                        in your server
                      </p>
                    </div>
                    <div>
                      <Input
                        id="webhook"
                        placeholder="https://discordapp.com/api/webhooks/1234/ABCD-EFG..."
                        disabled={!discordSettings.enabled}
                        value={discordSettings.webhook}
                        onChange={(e) =>
                          updateDiscordSettings({ webhook: e.target.value })
                        }
                        className="w-full text-ellipsis"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 items-center">
                    <div className="text-left">
                      <p className="text-sm text-muted-foreground font-bold">
                        Bot name
                      </p>
                    </div>
                    <div>
                      <Input
                        id="bot-name"
                        disabled={!discordSettings.enabled}
                        value={discordSettings.name}
                        onChange={(e) =>
                          updateDiscordSettings({ name: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleTestWebhook}
            className="mr-2"
          >
            <TestTubeIcon />
            Test
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
