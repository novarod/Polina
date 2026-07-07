"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { FormDialog } from "@/components/form/form-dialog";
import {
  MissionDialogFields,
  missionSchema,
  type MissionValues,
} from "@/components/missions/mission-dialog-fields";
import { Button } from "@/components/ui/button";
import { createMission } from "@/services/missions";

interface CreateMissionDialogProps {
  orgId: string;
  workspaceId: string;
}

export function CreateMissionDialog({
  orgId,
  workspaceId,
}: CreateMissionDialogProps) {
  const router = useRouter();
  const form = useForm<MissionValues>({
    resolver: zodResolver(missionSchema),
    defaultValues: { name: "", description: "" },
  });

  async function onSubmit(values: MissionValues) {
    await createMission(orgId, workspaceId, values);
    toast.success("Missão criada");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={<Button data-testid="create-mission">Nova missão</Button>}
      title="Nova missão"
      form={form}
      onSubmit={onSubmit}
      submitLabel="Criar"
    >
      <MissionDialogFields control={form.control} />
    </FormDialog>
  );
}
