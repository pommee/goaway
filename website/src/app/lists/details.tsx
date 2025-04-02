import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { ListEntry } from "@/pages/lists";
import {
  ArrowsClockwise,
  Eraser,
  Eye,
  ToggleLeft
} from "@phosphor-icons/react";

export function CardDetails(listEntry: ListEntry) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="bg-zinc-800 border-none hover:bg-zinc-700 text-white w-full mt-2 rounded-lg text-sm py-1 h-auto"
        >
          <Eye className="mr-2" size={16} />
          View Details
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-zinc-900 text-white border-zinc-800 rounded-xl w-1/3 max-w-none">
        <DialogHeader>
          <DialogTitle>{listEntry.name}</DialogTitle>
        </DialogHeader>
        <div className="flex gap-2 justify-between flex-wrap">
          <Button
            variant="outline"
            className="bg-zinc-800 border-none hover:bg-zinc-700 text-white flex-1 text-sm"
          >
            <ToggleLeft className="mr-1" size={16} />
            Toggle
          </Button>

          {listEntry.name !== "Custom" && (
            <Button
              variant="outline"
              className="bg-blue-600 border-none hover:bg-blue-500 text-white flex-1 text-sm"
            >
              <ArrowsClockwise className="mr-1" size={16} />
              Update
            </Button>
          )}

          <Button
            variant="outline"
            className="bg-red-600 border-none hover:bg-red-500 text-white flex-1 text-sm"
          >
            <Eraser className="mr-1" size={16} />
            Delete
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
