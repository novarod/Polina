"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { PencilIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { DeleteDialog } from "@/components/shared/delete-dialog";
import { CreateWorkspaceDialog } from "@/components/workspaces/create-workspace-dialog";
import { RenameWorkspaceDialog } from "@/components/workspaces/rename-workspace-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { roleAtLeast, type Role } from "@/lib/roles";
import { deleteWorkspace } from "@/services/workspaces";
import type { Workspace } from "@/types/workspace";

interface WorkspaceListProps {
  orgId: string;
  role: Role;
  workspaces: Workspace[];
}

export function WorkspaceList({ orgId, role, workspaces }: WorkspaceListProps) {
  const router = useRouter();
  const canEdit = roleAtLeast(role, "DESIGNER");

  return (
    <div className="grid gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-sm">Workspaces</h1>
        {canEdit && <CreateWorkspaceDialog orgId={orgId} />}
      </div>
      {workspaces.length === 0 ? (
        <p className="text-muted-foreground">
          {canEdit
            ? "Nenhum workspace ainda. Crie o primeiro para organizar suas missões."
            : "Nenhum workspace ainda."}
        </p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {workspaces.map((workspace) => (
            <Card key={workspace.id} data-testid="workspace-card">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <Link
                    href={`/orgs/${orgId}/workspaces/${workspace.id}`}
                    className="min-w-0"
                  >
                    <CardTitle className="truncate hover:text-primary">
                      {workspace.name}
                    </CardTitle>
                    {workspace.description && (
                      <CardDescription className="line-clamp-2">
                        {workspace.description}
                      </CardDescription>
                    )}
                  </Link>
                  {canEdit && (
                    <div className="flex shrink-0 gap-1">
                      <RenameWorkspaceDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Editar ${workspace.name}`}
                          >
                            <PencilIcon />
                          </Button>
                        }
                        workspace={workspace}
                      />
                      <DeleteDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Deletar ${workspace.name}`}
                          >
                            <Trash2Icon />
                          </Button>
                        }
                        entityLabel="o workspace"
                        name={workspace.name}
                        onConfirm={async () => {
                          await deleteWorkspace(orgId, workspace.id);
                          toast.success("Workspace deletado");
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
