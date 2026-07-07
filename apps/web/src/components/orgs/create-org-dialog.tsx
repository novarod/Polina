"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { FormDialog } from "@/components/form/form-dialog";
import { InputField } from "@/components/form/input-field";
import { Button } from "@/components/ui/button";
import { createOrganization } from "@/services/organizations";

const createOrgSchema = z.object({
  name: z
    .string()
    .min(2, "O nome precisa de pelo menos 2 caracteres")
    .max(255, "O nome pode ter no máximo 255 caracteres"),
  slug: z
    .string()
    .min(3, "O slug precisa de pelo menos 3 caracteres")
    .max(63, "O slug pode ter no máximo 63 caracteres")
    .regex(
      /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
      "Use apenas letras minúsculas, números e hífens (ex.: meu-estudio)"
    ),
});

type CreateOrgValues = z.infer<typeof createOrgSchema>;

export function CreateOrgDialog() {
  const router = useRouter();
  const form = useForm<CreateOrgValues>({
    resolver: zodResolver(createOrgSchema),
    defaultValues: { name: "", slug: "" },
  });

  async function onSubmit(values: CreateOrgValues) {
    await createOrganization(values);
    toast.success("Organização criada");
    router.refresh();
  }

  return (
    <FormDialog
      trigger={<Button data-testid="create-org">Nova organização</Button>}
      title="Nova organização"
      description="O slug identifica a organização e não pode ser alterado depois."
      form={form}
      onSubmit={onSubmit}
      submitLabel="Criar"
    >
      <InputField control={form.control} name="name" label="Nome" />
      <InputField
        control={form.control}
        name="slug"
        label="Slug"
        placeholder="meu-estudio"
      />
    </FormDialog>
  );
}
