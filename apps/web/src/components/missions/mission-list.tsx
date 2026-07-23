"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { PencilIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { CreateMissionDialog } from "@/components/missions/create-mission-dialog";
import { MissionStatusBadge } from "@/components/missions/mission-status-badge";
import { RenameMissionDialog } from "@/components/missions/rename-mission-dialog";
import { DeleteDialog } from "@/components/shared/delete-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useOrgStatus } from "@/hooks/use-org-status";
import { roleAtLeast, type Role } from "@/lib/roles";
import { deleteMission } from "@/services/missions";
import type { Mission } from "@/types/mission";

interface MissionListProps {
  orgId: string;
  workspaceId: string;
  role: Role;
  missions: Mission[];
}

export function MissionList({
  orgId,
  workspaceId,
  role,
  missions,
}: MissionListProps) {
  const router = useRouter();
  const canEdit = roleAtLeast(role, "DESIGNER");
  const editingCounts = useOrgStatus(orgId);

  return (
    <div className="grid gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-sm">Missões</h1>
        {canEdit && (
          <CreateMissionDialog orgId={orgId} workspaceId={workspaceId} />
        )}
      </div>
      {missions.length === 0 ? (
        <p className="text-muted-foreground">
          {canEdit
            ? "Nenhuma missão ainda. Crie a primeira para começar a desenhar."
            : "Nenhuma missão ainda."}
        </p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {missions.map((mission) => (
            <Card key={mission.id} data-testid="mission-card">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <Link
                    href={`/orgs/${orgId}/workspaces/${workspaceId}/missions/${mission.id}`}
                    className="grid min-w-0 gap-1"
                  >
                    <CardTitle className="truncate hover:text-primary">
                      {mission.name}
                    </CardTitle>
                    <div className="flex items-center gap-2">
                      <MissionStatusBadge status={mission.status} />
                      {(editingCounts[mission.id] ?? 0) > 0 && (
                        <span
                          data-testid="editing-badge"
                          className="rounded-sm border-2 border-primary px-1.5 py-0.5 font-display text-[8px] text-primary"
                        >
                          {editingCounts[mission.id]} editando
                        </span>
                      )}
                    </div>
                    {mission.description && (
                      <CardDescription className="line-clamp-2">
                        {mission.description}
                      </CardDescription>
                    )}
                  </Link>
                  {canEdit && (
                    <div className="flex shrink-0 gap-1">
                      <RenameMissionDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Editar ${mission.name}`}
                          >
                            <PencilIcon />
                          </Button>
                        }
                        mission={mission}
                      />
                      <DeleteDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Deletar ${mission.name}`}
                          >
                            <Trash2Icon />
                          </Button>
                        }
                        entityLabel="a missão"
                        name={mission.name}
                        onConfirm={async () => {
                          await deleteMission(orgId, workspaceId, mission.id);
                          toast.success("Missão deletada");
                          router.refresh();
                        }}
                      />
                    </div>
                  )}
                </div>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
