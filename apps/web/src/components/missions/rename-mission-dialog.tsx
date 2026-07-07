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
import { updateMission } from "@/services/missions";
import type { Mission } from "@/types/mission";

interface RenameMissionDialogProps {
  trigger: React.ReactNode;
  mission: Mission;
}

export function RenameMissionDialog({
  trigger,
  mission,
}: RenameMissionDialogProps) {
  const router = useRouter();
  const form = useForm<MissionValues>({
    resolver: zodResolver(missionSchema),
    defaultValues: { name: mission.name, description: mission.description },
  });

  async function onSubmit(values: MissionValues) {
    await updateMission(
      mission.organization_id,
      mission.workspace_id,
      mission.id,
      values
    );
    toast.success("Missão atualizada");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={trigger}
      title="Editar missão"
      form={form}
      onSubmit={onSubmit}
      submitLabel="Salvar"
    >
      <MissionDialogFields control={form.control} />
    </FormDialog>
  );
}
