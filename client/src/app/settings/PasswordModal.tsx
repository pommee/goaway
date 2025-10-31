import { WarningIcon } from "@phosphor-icons/react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export const PasswordModal = ({
  open,
  onClose,
  onSubmit,
  passwords,
  setPasswords,
  error,
  setError
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: () => void;
  passwords: { current: string; new: string; confirm: string };
  setPasswords: (p: { current: string; new: string; confirm: string }) => void;
  error: string;
  setError: (e: string) => void;
}) => (
  <Dialog open={open} onOpenChange={onClose}>
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Change Password</DialogTitle>
        <DialogDescription>
          Update your password. You'll be logged out after changing it.
        </DialogDescription>
      </DialogHeader>

      {error && (
        <div className="flex items-center bg-red-900/20 text-red-500 p-2 rounded text-sm">
          <WarningIcon className="mr-2" />
          {error}
        </div>
      )}

      <div className="space-y-4">
        {["current", "new", "confirm"].map((type) => (
          <div key={type} className="space-y-2">
            <label className="text-sm font-medium">
              {type === "current"
                ? "Current Password"
                : type === "new"
                  ? "New Password"
                  : "Confirm Password"}
            </label>
            <Input
              type="password"
              value={passwords[type as keyof typeof passwords]}
              onChange={(e) => {
                setPasswords({ ...passwords, [type]: e.target.value });
                setError("");
              }}
              placeholder={`Enter ${type} password`}
            />
          </div>
        ))}
      </div>

      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          Cancel
        </Button>
        <Button onClick={onSubmit}>Update</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);
