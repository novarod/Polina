import { randomUUID } from "node:crypto";

import { expect, test } from "../fixtures/session";
import { MissionPage } from "../pages/mission-page";
import { OrgPage } from "../pages/org-page";
import { OrgsPage } from "../pages/orgs-page";
import { WorkspacePage } from "../pages/workspace-page";

test.beforeEach(async ({ context, page, sessionToken }) => {
  await context.addCookies([
    {
      name: "session",
      value: sessionToken,
      url: "http://localhost:3000",
      httpOnly: true,
      sameSite: "Strict",
    },
  ]);
  await page.goto("/orgs");
  await expect(page).toHaveURL("/orgs");
});

test("fluxo completo: criar, navegar, renomear e deletar org → workspace → missão", async ({
  page,
}) => {
  const stamp = Date.now();
  const orgName = `Estúdio E2E ${stamp}`;
  const orgsPage = new OrgsPage(page);
  const orgPage = new OrgPage(page);
  const workspacePage = new WorkspacePage(page);
  const missionPage = new MissionPage(page);

  await orgsPage.createOrg(orgName, `estudio-e2e-${stamp}`);
  await expect(orgsPage.orgCard(orgName)).toBeVisible();

  await orgsPage.openOrg(orgName);
  await expect(orgPage.heading).toBeVisible();
  await expect(orgPage.breadcrumb).toContainText(orgName);

  await orgPage.createWorkspace("DLC de Natal", "Missões do evento");
  await expect(orgPage.workspaceCard("DLC de Natal")).toBeVisible();

  await orgPage.openWorkspace("DLC de Natal");
  await expect(workspacePage.heading).toBeVisible();
  await expect(workspacePage.breadcrumb).toContainText("DLC de Natal");

  await workspacePage.createMission("Resgate", "Salvar o aldeão");
  await expect(workspacePage.missionCard("Resgate")).toBeVisible();

  await workspacePage.openMission("Resgate");
  await expect(missionPage.editorCanvas).toBeVisible();
  await expect(missionPage.statusBadge).toHaveText("Rascunho");
  const missionUrl = page.url();

  await page.goto(missionUrl);
  await expect(missionPage.editorCanvas).toBeVisible();

  await missionPage.breadcrumb
    .getByRole("link", { name: "DLC de Natal" })
    .click();
  await expect(workspacePage.heading).toBeVisible();

  await workspacePage.renameMission("Resgate", "Resgate na Vila");
  await expect(workspacePage.missionCard("Resgate na Vila")).toBeVisible();

  await workspacePage.deleteMission("Resgate na Vila");
  await expect(page.getByText("Nenhuma missão ainda", { exact: false })).toBeVisible();

  await workspacePage.breadcrumb.getByRole("link", { name: orgName }).click();
  await orgPage.deleteWorkspace("DLC de Natal");
  await expect(
    page.getByText("Nenhum workspace ainda", { exact: false })
  ).toBeVisible();

  await orgPage.breadcrumb
    .getByRole("link", { name: "Organizações" })
    .click();
  await orgsPage.deleteOrg(orgName);
  await expect(orgsPage.orgCard(orgName)).toHaveCount(0);
});

test("org inexistente na URL mostra a página 404", async ({ page }) => {
  await page.goto(`/orgs/${randomUUID()}`);

  await expect(page.getByText("Página não encontrada.")).toBeVisible();
});

test("slug duplicado mostra erro dentro do dialog", async ({ page }) => {
  const stamp = Date.now();
  const orgsPage = new OrgsPage(page);

  await orgsPage.createOrg(`Org Dup ${stamp}`, `org-dup-${stamp}`);
  await expect(orgsPage.orgCard(`Org Dup ${stamp}`)).toBeVisible();

  await orgsPage.createOrg(`Org Dup B ${stamp}`, `org-dup-${stamp}`);
  await expect(page.getByTestId("dialog-error")).toBeVisible();
  await page.keyboard.press("Escape");

  await orgsPage.deleteOrg(`Org Dup ${stamp}`);
  await expect(orgsPage.orgCard(`Org Dup ${stamp}`)).toHaveCount(0);
});
