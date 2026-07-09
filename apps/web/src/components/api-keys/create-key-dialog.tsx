"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { CopyIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { FormDialog } from "@/components/form/form-dialog";
import { InputField } from "@/components/form/input-field";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { createApiKey } from "@/services/api-keys";
import type { CreatedApiKey } from "@/types/api-key";

const createKeySchema = z.object({
  name: z
    .string()
    .min(1, "Informe um nome para a chave")
    .max(255, "O nome pode ter no máximo 255 caracteres"),
});

type CreateKeyValues = z.infer<typeof createKeySchema>;

export function CreateKeyDialog({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [createdKey, setCreatedKey] = useState<CreatedApiKey | null>(null);
  const form = useForm<CreateKeyValues>({
    resolver: zodResolver(createKeySchema),
    defaultValues: { name: "" },
  });

  async function onSubmit(values: CreateKeyValues) {
    const key = await createApiKey(orgId, values.name);
    setCreatedKey(key);
    router.refresh();
  }

  async function copySecret() {
    if (!createdKey) {
      return;
    }
    try {
      await navigator.clipboard.writeText(createdKey.key);
      toast.success("Chave copiada");
    } catch {
      toast.error("Não foi possível copiar");
    }
  }

  return (
    <>
      <FormDialog
        trigger={<Button data-testid="create-key">Nova chave</Button>}
        title="Nova API key"
        description="A chave dá acesso de engine à organização inteira."
        form={form}
        onSubmit={onSubmit}
        submitLabel="Criar"
      >
        <InputField control={form.control} name="name" label="Nome" />
      </FormDialog>
      <Dialog
        open={createdKey !== null}
        onOpenChange={(open) => {
          if (!open) {
            setCreatedKey(null);
          }
        }}
      >
        <DialogContent data-testid="key-secret-dialog">
          <DialogHeader>
            <DialogTitle>Chave criada</DialogTitle>
            <DialogDescription>
              Copie agora — o segredo não será exibido novamente.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2">
            <code
              data-testid="key-secret"
              className="min-w-0 flex-1 truncate rounded-sm bg-muted p-2 text-xs"
            >
              {createdKey?.key}
            </code>
            <Button
              variant="outline"
              size="icon-sm"
              aria-label="Copiar chave"
              onClick={copySecret}
            >
              <CopyIcon />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => setCreatedKey(null)}>Fechar</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
