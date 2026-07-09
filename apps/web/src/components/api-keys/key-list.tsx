"use client";

import { useRouter } from "next/navigation";
import { Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { DeleteDialog } from "@/components/shared/delete-dialog";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { revokeApiKey } from "@/services/api-keys";
import type { ApiKey } from "@/types/api-key";

export function KeyList({ orgId, keys }: { orgId: string; keys: ApiKey[] }) {
  const router = useRouter();

  if (keys.length === 0) {
    return (
      <p className="text-muted-foreground">
        Nenhuma chave ainda. Crie uma para conectar o plugin.
      </p>
    );
  }

  return (
    <ul className="grid gap-2">
      {keys.map((apiKey) => {
        const revoked = apiKey.revoked_at !== null;
        return (
          <li
            key={apiKey.id}
            data-testid="key-row"
            className={cn(
              "flex items-center gap-3 rounded-sm border-2 border-foreground/40 bg-card px-3 py-2",
              revoked && "opacity-50"
            )}
          >
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{apiKey.name}</p>
              <p className="text-xs text-muted-foreground">
                Criada em {apiKey.created_at.slice(0, 10)} · Último uso:{" "}
                {apiKey.last_used_at
                  ? apiKey.last_used_at.slice(0, 10)
                  : "nunca"}
              </p>
            </div>
            {revoked ? (
              <span
                data-testid="key-revoked"
                className="rounded-sm bg-muted px-2 py-0.5 text-xs text-muted-foreground"
              >
                Revogada
              </span>
            ) : (
              <DeleteDialog
                trigger={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Revogar ${apiKey.name}`}
                  >
                    <Trash2Icon />
                  </Button>
                }
                entityLabel="a chave"
                name={apiKey.name}
                actionLabel="Revogar"
                pendingLabel="Revogando..."
                onConfirm={async () => {
                  await revokeApiKey(orgId, apiKey.id);
                  toast.success("Chave revogada");
                  router.refresh();
                }}
              />
            )}
          </li>
        );
      })}
    </ul>
  );
}
