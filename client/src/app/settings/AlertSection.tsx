import { Button } from "@/components/ui/button";
import { SettingRow } from "./SettingsRow";

export const AlertsSection = ({
  onConfigureClick
}: {
  onConfigureClick: () => void;
}) => (
  <SettingRow
    title="Configure"
    description="Set up how you receive notifications for important events."
    action={
      <Button variant="outline" onClick={onConfigureClick}>
        Open
      </Button>
    }
  />
);
