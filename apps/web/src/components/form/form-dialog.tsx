"use client";

import { useState } from "react";
import type { FieldValues, UseFormReturn } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Form } from "@/components/ui/form";
import { ApiError } from "@/services/api";

interface FormDialogProps<TFieldValues extends FieldValues> {
  trigger: React.ReactNode;
  title: string;
  description?: string;
  form: UseFormReturn<TFieldValues>;
  onSubmit: (values: TFieldValues) => Promise<void>;
  submitLabel: string;
  children: React.ReactNode;
}

export function FormDialog<TFieldValues extends FieldValues>({
  trigger,
  title,
  description,
  form,
  onSubmit,
  submitLabel,
  children,
}: FormDialogProps<TFieldValues>) {
  const [open, setOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  function onOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      form.reset();
      setSubmitError(null);
    }
  }

  async function handleSubmit(values: TFieldValues) {
    setSubmitError(null);
    try {
      await onSubmit(values);
      setOpen(false);
      form.reset();
    } catch (error) {
      setSubmitError(
        error instanceof ApiError
          ? error.message
          : "Algo deu errado, tente novamente"
      );
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="grid gap-4"
            noValidate
          >
            {children}
            {submitError && (
              <p
                role="alert"
                data-testid="dialog-error"
                className="text-sm font-medium text-destructive"
              >
                {submitError}
              </p>
            )}
            <DialogFooter>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting ? "Salvando..." : submitLabel}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
