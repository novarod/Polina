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
  for (let attempt = 1; attempt <= 6; attempt++) {
    const response = await fetch(`${apiUrl}/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(account),
    });
    if (response.ok) {
      return account;
    }
    if (response.status !== 429) {
      throw new Error(
        `Falha ao registrar usuário de teste na API: status ${response.status}`
      );
    }
    await new Promise((resolve) =>
      setTimeout(resolve, 13000 + workerIndex * 500)
    );
  }
  throw new Error(
    "Falha ao registrar usuário de teste na API: rate limit persistente (429)"
  );
}

async function loginSessionToken(account: Account): Promise<string> {
  for (let attempt = 1; attempt <= 6; attempt++) {
    const response = await fetch(`${apiUrl}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: account.email,
        password: account.password,
      }),
    });
    if (response.ok) {
      const setCookie = response.headers.get("set-cookie") ?? "";
      const match = /session=([^;]+)/.exec(setCookie);
      if (!match) {
        throw new Error("Login de teste não retornou o cookie de sessão");
      }
      return match[1];
    }
    if (response.status !== 429) {
      throw new Error(
        `Falha no login do usuário de teste: status ${response.status}`
      );
    }
    await new Promise((resolve) => setTimeout(resolve, 13000));
  }
  throw new Error("Falha no login do usuário de teste: rate limit (429)");
}

interface SessionWorkerFixtures {
  account: Account;
  sessionToken: string;
}

export const test = base.extend<Record<never, never>, SessionWorkerFixtures>({
  account: [
    async ({}, use, workerInfo) => {
      await assertApiIsUp();
      const account = await registerAccount(workerInfo.workerIndex);
      await use(account);
    },
    { scope: "worker" },
  ],
  sessionToken: [
    async ({ account }, use) => {
      await use(await loginSessionToken(account));
    },
    { scope: "worker" },
  ],
});

export { expect };
