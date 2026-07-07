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
import { Button } from "@/components/ui/button";
import { createWorkspace } from "@/services/workspaces";

export function CreateWorkspaceDialog({ orgId }: { orgId: string }) {
  const router = useRouter();
  const form = useForm<WorkspaceValues>({
    resolver: zodResolver(workspaceSchema),
    defaultValues: { name: "", description: "" },
  });

  async function onSubmit(values: WorkspaceValues) {
    await createWorkspace(orgId, values);
    toast.success("Workspace criado");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={<Button data-testid="create-workspace">Novo workspace</Button>}
      title="Novo workspace"
      form={form}
      onSubmit={onSubmit}
      submitLabel="Criar"
    >
      <WorkspaceDialogFields control={form.control} />
    </FormDialog>
  );
}
