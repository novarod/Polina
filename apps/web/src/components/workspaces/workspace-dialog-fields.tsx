"use client";

import type { Control } from "react-hook-form";
import { z } from "zod";

import { InputField } from "@/components/form/input-field";
import { TextareaField } from "@/components/form/textarea-field";

export const workspaceSchema = z.object({
  name: z
    .string()
    .min(2, "O nome precisa de pelo menos 2 caracteres")
    .max(255, "O nome pode ter no máximo 255 caracteres"),
  description: z
    .string()
    .max(1000, "A descrição pode ter no máximo 1000 caracteres"),
});

export type WorkspaceValues = z.infer<typeof workspaceSchema>;

export function WorkspaceDialogFields({
  control,
}: {
  control: Control<WorkspaceValues>;
}) {
  return (
    <>
      <InputField control={control} name="name" label="Nome" />
      <TextareaField
        control={control}
        name="description"
        label="Descrição (opcional)"
      />
    </>
  );
}
