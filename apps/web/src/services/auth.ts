import { apiFetch } from "@/services/api";
import type { SessionUser } from "@/types/auth";

export interface LoginInput {
  email: string;
  password: string;
}

export function login(input: LoginInput): Promise<SessionUser> {
  return apiFetch<SessionUser>("/auth/login", {
    method: "POST",
    body: input,
    redirectOn401: false,
  });
}

export function logout(): Promise<void> {
  return apiFetch<void>("/auth/logout", { method: "POST" });
}
