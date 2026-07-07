"use client";

import type * as React from "react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";

type TextareaFieldProps<TFieldValues extends FieldValues> = {
  control: Control<TFieldValues>;
  name: FieldPath<TFieldValues>;
  label: string;
} & Omit<React.ComponentProps<typeof Textarea>, "name">;

export function TextareaField<TFieldValues extends FieldValues>({
  control,
  name,
  label,
  ...textareaProps
}: TextareaFieldProps<TFieldValues>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Textarea {...textareaProps} {...field} />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
