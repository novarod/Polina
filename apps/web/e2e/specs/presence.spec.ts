import type { BrowserContext, Page } from "@playwright/test";

import { expect, test } from "../fixtures/session";
import { seedMission } from "../fixtures/seed";

const simpleGraph = {
  nodes: [
    { id: "inicio", type: "START" },
    { id: "fim", type: "END" },
  ],
  edges: [{ id: "e1", source: "inicio", target: "fim" }],
};

async function withSession(
  context: BrowserContext,
  sessionToken: string
): Promise<void> {
  await context.addCookies([
    {
      name: "session",
      value: sessionToken,
      url: "http://localhost:3000",
      httpOnly: true,
      sameSite: "Strict",
    },
  ]);
}

test.beforeEach(async ({ context, sessionToken }) => {
  await withSession(context, sessionToken);
});

async function openSecondSession(
  browserContext: () => Promise<BrowserContext>,
  sessionToken: string
): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browserContext();
  await withSession(context, sessionToken);
  const page = await context.newPage();
  return { context, page };
}

test("presença própria aparece como avatar na mission", async ({
  page,
  sessionToken,
  account,
}) => {
  const seeded = await seedMission(sessionToken, simpleGraph);
  await page.goto(seeded.path);

  await expect(page.getByTestId("presence-avatars")).toBeVisible();
  await expect(page.getByTestId("presence-avatar")).toHaveCount(1);
  await expect(page.getByTestId("presence-avatar")).toHaveAttribute(
    "title",
    account.name
  );
});

test("badge 'editando' aparece na lista, deduplica abas e some ao sair", async ({
  page,
  browser,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, simpleGraph);
  const listPath = `/orgs/${seeded.orgId}/workspaces/${seeded.workspaceId}`;

  await page.goto(listPath);
  await expect(page.getByTestId("mission-card")).toBeVisible();
  await expect(page.getByTestId("editing-badge")).toHaveCount(0);

  const editorA = await openSecondSession(
    () => browser.newContext(),
    sessionToken
  );
  await editorA.page.goto(seeded.path);
  await expect(editorA.page.getByTestId("mission-editor")).toBeVisible();

  await expect(page.getByTestId("editing-badge")).toHaveText("1 editando");

  const editorB = await openSecondSession(
    () => browser.newContext(),
    sessionToken
  );
  await editorB.page.goto(seeded.path);
  await expect(editorB.page.getByTestId("mission-editor")).toBeVisible();

  await editorB.page.waitForTimeout(1000);
  await expect(page.getByTestId("editing-badge")).toHaveText("1 editando");

  await editorA.context.close();
  await editorB.context.close();
  await expect(page.getByTestId("editing-badge")).toHaveCount(0, {
    timeout: 15_000,
  });
});

test("presença deduplicada: duas abas do mesmo usuário mostram um avatar só", async ({
  page,
  browser,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, simpleGraph);
  await page.goto(seeded.path);
  await expect(page.getByTestId("presence-avatar")).toHaveCount(1);

  const second = await openSecondSession(
    () => browser.newContext(),
    sessionToken
  );
  await second.page.goto(seeded.path);
  await expect(second.page.getByTestId("presence-avatar")).toHaveCount(1);

  await second.page.waitForTimeout(1000);
  await expect(page.getByTestId("presence-avatar")).toHaveCount(1);

  await second.context.close();
  await expect(page.getByTestId("presence-avatar")).toHaveCount(1);
});
