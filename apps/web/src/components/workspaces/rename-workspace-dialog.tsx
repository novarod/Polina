"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { FormDialog } from "@/components/form/form-dialog";
import {
  WorkspaceDialogFields,
  workspaceSchema,
  type WorkspaceValues,
} from "@/components/workspaces/workspace-dialog-fields";
import { updateWorkspace } from "@/services/workspaces";
import type { Workspace } from "@/types/workspace";

interface RenameWorkspaceDialogProps {
  trigger: React.ReactNode;
  workspace: Workspace;
}

export function RenameWorkspaceDialog({
  trigger,
  workspace,
}: RenameWorkspaceDialogProps) {
  const router = useRouter();
  const form = useForm<WorkspaceValues>({
    resolver: zodResolver(workspaceSchema),
    defaultValues: {
      name: workspace.name,
      description: workspace.description,
    },
  });

  async function onSubmit(values: WorkspaceValues) {
    await updateWorkspace(workspace.organization_id, workspace.id, values);
    toast.success("Workspace atualizado");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={trigger}
      title="Editar workspace"
      form={form}
      onSubmit={onSubmit}
      submitLabel="Salvar"
    >
      <WorkspaceDialogFields control={form.control} />
    </FormDialog>
  );
}
