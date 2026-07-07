"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { FormDialog } from "@/components/form/form-dialog";
import { InputField } from "@/components/form/input-field";
import { renameOrganization } from "@/services/organizations";

const renameOrgSchema = z.object({
  name: z
    .string()
    .min(2, "O nome precisa de pelo menos 2 caracteres")
    .max(255, "O nome pode ter no máximo 255 caracteres"),
});

type RenameOrgValues = z.infer<typeof renameOrgSchema>;

interface RenameOrgDialogProps {
  trigger: React.ReactNode;
  orgId: string;
  currentName: string;
}

export function RenameOrgDialog({
  trigger,
  orgId,
  currentName,
}: RenameOrgDialogProps) {
  const router = useRouter();
  const form = useForm<RenameOrgValues>({
    resolver: zodResolver(renameOrgSchema),
    defaultValues: { name: currentName },
  });

  async function onSubmit(values: RenameOrgValues) {
    await renameOrganization(orgId, values.name);
    toast.success("Organização renomeada");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={trigger}
      title="Renomear organização"
      form={form}
      onSubmit={onSubmit}
      submitLabel="Salvar"
    >
      <InputField control={form.control} name="name" label="Nome" />
    </FormDialog>
  );
}
