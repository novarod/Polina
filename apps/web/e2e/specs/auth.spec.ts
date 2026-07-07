import type { Page } from "@playwright/test";

import { expect, test, type Account } from "../fixtures/session";
import { LoginPage } from "../pages/login-page";
import { OrgsPage } from "../pages/orgs-page";

const RATE_LIMIT_WAIT_MS = 13_000;

async function loginUntilBudget(
  page: Page,
  loginPage: LoginPage,
  account: Account,
  password: string,
  expectedError?: string
): Promise<void> {
  for (let attempt = 1; attempt <= 5; attempt++) {
    await loginPage.login(account.email, password);
    if (!expectedError) {
      try {
        await expect(page).toHaveURL("/orgs", { timeout: 4000 });
        return;
      } catch {
        const message = await loginPage.errorAlert
          .textContent()
          .catch(() => null);
        if (!message?.includes("rate limit")) {
          throw new Error(`Login de UI falhou: ${message}`);
        }
      }
    } else {
      const message = await loginPage.errorAlert.textContent();
      if (message === expectedError) {
        return;
      }
      if (!message?.includes("rate limit")) {
        throw new Error(`Erro inesperado no login de UI: ${message}`);
      }
    }
    await page.waitForTimeout(RATE_LIMIT_WAIT_MS);
  }
  throw new Error("Login de UI esbarrando no rate limit persistentemente");
}

async function injectSession(
  page: Page,
  sessionToken: string
): Promise<void> {
  await page.context().addCookies([
    {
      name: "session",
      value: sessionToken,
      url: "http://localhost:3000",
      httpOnly: true,
      sameSite: "Strict",
    },
  ]);
}

test("login com credenciais válidas leva às organizações com o nome do usuário", async ({
  page,
  account,
}) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginUntilBudget(page, loginPage, account, account.password);

  const orgsPage = new OrgsPage(page);
  await expect(page).toHaveURL("/orgs");
  await expect(orgsPage.userMenu).toContainText(account.name);
});

test("credenciais inválidas mostram erro e permanecem em /login", async ({
  page,
  account,
}) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginUntilBudget(
    page,
    loginPage,
    account,
    "senha-errada-123",
    "Email ou senha inválidos"
  );

  await expect(page).toHaveURL("/login");
});

test("acessar /orgs sem sessão redireciona para /login", async ({ page }) => {
  const orgsPage = new OrgsPage(page);
  await orgsPage.goto();

  await expect(page).toHaveURL("/login");
});

test("reload em /orgs mantém a sessão", async ({
  page,
  account,
  sessionToken,
}) => {
  await injectSession(page, sessionToken);
  await page.goto("/orgs");
  await expect(page).toHaveURL("/orgs");

  await page.reload();

  const orgsPage = new OrgsPage(page);
  await expect(page).toHaveURL("/orgs");
  await expect(orgsPage.userMenu).toContainText(account.name);
});

test("logout encerra a sessão e /orgs volta a redirecionar", async ({
  page,
  sessionToken,
}) => {
  await injectSession(page, sessionToken);
  const orgsPage = new OrgsPage(page);
  await orgsPage.goto();
  await expect(page).toHaveURL("/orgs");

  await orgsPage.logout();
  await expect(page).toHaveURL("/login");

  await orgsPage.goto();
  await expect(page).toHaveURL("/login");
});
