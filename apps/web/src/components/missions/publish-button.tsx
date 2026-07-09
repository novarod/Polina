"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/services/api";
import { publishMission } from "@/services/missions";

interface PublishButtonProps {
  orgId: string;
  workspaceId: string;
  missionId: string;
  activeHash: string | null;
  dirty: boolean;
}

export function PublishButton({
  orgId,
  workspaceId,
  missionId,
  activeHash,
  dirty,
}: PublishButtonProps) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function onOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setError(null);
    }
  }

  async function confirm(event: React.MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    setPublishing(true);
    setError(null);
    try {
      const result = await publishMission(orgId, workspaceId, missionId);
      if (result.hash === activeHash) {
        toast.info("Nada mudou — o conteúdo é idêntico à versão ativa");
      } else {
        toast.success(`Versão v${result.version} publicada`);
      }
      setOpen(false);
      router.refresh();
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Algo deu errado ao publicar, tente novamente"
      );
    } finally {
      setPublishing(false);
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogTrigger asChild>
        <Button
          size="sm"
          data-testid="publish-button"
          disabled={dirty}
          title={dirty ? "Salve o grafo antes de publicar" : undefined}
        >
          Publicar
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Publicar missão?</AlertDialogTitle>
          <AlertDialogDescription>
            A versão publicada passa a ser servida ao plugin imediatamente.
          </AlertDialogDescription>
        </AlertDialogHeader>
        {error && (
          <p
            role="alert"
            data-testid="publish-error"
            className="text-sm font-medium text-destructive"
          >
            {error}
          </p>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={publishing}>Cancelar</AlertDialogCancel>
          <AlertDialogAction
            disabled={publishing}
            onClick={confirm}
            data-testid="confirm-publish"
          >
            {publishing ? "Publicando..." : "Publicar"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
