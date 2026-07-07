import { test as base, expect } from "@playwright/test";

const apiUrl = process.env.API_URL ?? "http://localhost:8080";

export interface Account {
  name: string;
  email: string;
  password: string;
}

async function assertApiIsUp(): Promise<void> {
  try {
    const response = await fetch(`${apiUrl}/health`);
    if (!response.ok) {
      throw new Error(`status ${response.status}`);
    }
  } catch (error) {
    throw new Error(
      `A API precisa estar no ar em ${apiUrl} para rodar o e2e (docker compose up). Motivo: ${String(error)}`
    );
  }
}

async function registerAccount(workerIndex: number): Promise<Account> {
  const account: Account = {
    name: "E2E Tester",
    email: `e2e-${Date.now()}-${workerIndex}@polina.test`,
    password: "correct-horse-battery",
  };
  const response = await fetch(`${apiUrl}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(account),
  });
  if (!response.ok) {
    throw new Error(
      `Falha ao registrar usuário de teste na API: status ${response.status}`
    );
  }
  return account;
}

export const test = base.extend<Record<never, never>, { account: Account }>({
  account: [
    async ({}, use, workerInfo) => {
      await assertApiIsUp();
      const account = await registerAccount(workerInfo.workerIndex);
      await use(account);
    },
    { scope: "worker" },
  ],
});

export { expect };
