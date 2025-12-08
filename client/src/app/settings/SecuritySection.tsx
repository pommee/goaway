import { Button } from "@/components/ui/button";
import { SettingRow } from "./SettingsRow";

export const SecuritySection = ({
  onPasswordClick,
  onApiKeyClick
}: {
  onPasswordClick: () => void;
  onApiKeyClick: () => void;
}) => (
  <div className="space-y-2">
    <SettingRow
      title="Change password"
      description="Update dashboard login password."
      action={
        <Button variant="outline" onClick={onPasswordClick}>
          Change Password
        </Button>
      }
    />
    <SettingRow
      title="API Keys"
      description="Manage programmatic access keys."
      action={
        <Button variant="outline" onClick={onApiKeyClick}>
          Manage Keys
        </Button>
      }
    />
  </div>
);
