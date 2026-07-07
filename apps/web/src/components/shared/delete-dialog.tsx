"use client";

import { useState } from "react";

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
import { ApiError } from "@/services/api";

interface DeleteDialogProps {
  trigger: React.ReactNode;
  entityLabel: string;
  name: string;
  onConfirm: () => Promise<void>;
}

export function DeleteDialog({
  trigger,
  entityLabel,
  name,
  onConfirm,
}: DeleteDialogProps) {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function onOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setError(null);
    }
  }

  async function handleConfirm(event: React.MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    setPending(true);
    setError(null);
    try {
      await onConfirm();
      setOpen(false);
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Algo deu errado, tente novamente"
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogTrigger asChild>{trigger}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Deletar {entityLabel} “{name}”?
          </AlertDialogTitle>
          <AlertDialogDescription>
            Isso não pode ser desfeito.
          </AlertDialogDescription>
        </AlertDialogHeader>
        {error && (
          <p
            role="alert"
            data-testid="dialog-error"
            className="text-sm font-medium text-destructive"
          >
            {error}
          </p>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>Cancelar</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={pending}
            onClick={handleConfirm}
          >
            {pending ? "Deletando..." : "Deletar"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
