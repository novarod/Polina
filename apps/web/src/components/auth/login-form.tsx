"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { InputField } from "@/components/form/input-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { ApiError } from "@/services/api";
import { login } from "@/services/auth";

const loginSchema = z.object({
  email: z.email("Informe um email válido"),
  password: z.string().min(1, "Informe a senha"),
});

type LoginValues = z.infer<typeof loginSchema>;

export function LoginForm() {
  const router = useRouter();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  async function onSubmit(values: LoginValues) {
    setSubmitError(null);
    try {
      await login(values);
      router.push("/orgs");
      router.refresh();
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setSubmitError("Email ou senha inválidos");
        return;
      }
      setSubmitError(
        error instanceof ApiError
          ? error.message
          : "Algo deu errado, tente novamente"
      );
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="grid gap-4"
        noValidate
      >
        <InputField
          control={form.control}
          name="email"
          label="Email"
          type="email"
          placeholder="voce@estudio.com"
          autoComplete="email"
        />
        <InputField
          control={form.control}
          name="password"
          label="Senha"
          type="password"
          autoComplete="current-password"
        />
        {submitError && (
          <p
            role="alert"
            data-testid="login-error"
            className="text-sm font-medium text-destructive"
          >
            {submitError}
          </p>
        )}
        <Button type="submit" disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting ? "Entrando..." : "Entrar"}
        </Button>
      </form>
    </Form>
  );
}
