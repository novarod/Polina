import { cn } from "@/lib/utils";
import type { MissionStatus } from "@/types/mission";

const statusLabels: Record<MissionStatus, string> = {
  DRAFT: "Rascunho",
  APPROVED: "Publicada",
};

export function MissionStatusBadge({ status }: { status: MissionStatus }) {
  return (
    <span
      data-testid="mission-status"
      className={cn(
        "inline-flex w-fit items-center rounded-sm px-2 py-0.5 text-xs font-medium",
        status === "APPROVED"
          ? "bg-primary/15 text-primary"
          : "bg-muted text-muted-foreground"
      )}
    >
      {statusLabels[status]}
    </span>
  );
}
