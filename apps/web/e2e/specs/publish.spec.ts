import { expect, test } from "../fixtures/session";
import { seedMission } from "../fixtures/seed";
import { CanvasPage } from "../pages/canvas-page";

const apiUrl = process.env.API_URL ?? "http://localhost:8080";

const linearGraph = {
  nodes: [
    { id: "inicio", type: "START", position: { x: 0, y: 0 } },
    {
      id: "conversa",
      type: "DIALOGUE",
      data: { npc: "Aldeão" },
      position: { x: 0, y: 160 },
    },
    { id: "fim", type: "END", position: { x: 300, y: 320 } },
  ],
  edges: [
    { id: "e1", source: "inicio", target: "conversa" },
    { id: "e2", source: "conversa", target: "fim" },
  ],
};

async function engineHash(
  missionId: string,
  key: string
): Promise<{ status: number; hash?: string }> {
  const response = await fetch(
    `${apiUrl}/engine/missions/${missionId}/active/hash`,
    { headers: { "x-api-key": key } }
  );
  if (!response.ok) {
    return { status: response.status };
  }
  const body = (await response.json()) as { hash: string };
  return { status: response.status, hash: body.hash };
}

test.beforeEach(async ({ context, sessionToken }) => {
  await context.addCookies([
    {
      name: "session",
      value: sessionToken,
      httpOnly: true,
      sameSite: "Strict",
      url: "http://localhost:3000",
    },
  ]);
});

test("o pitch: publicar na UI e o engine detectar o hash e a mudança dele", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, linearGraph);
  const canvas = new CanvasPage(page);

  await page.goto(seeded.path);
  await expect(page.getByTestId("version-list")).toContainText(
    "Nenhuma versão publicada ainda."
  );

  await page.getByTestId("publish-button").click();
  await page.getByTestId("confirm-publish").click();
  await expect(page.getByTestId("mission-status")).toHaveText("Publicada");
  await expect(page.getByTestId("version-item")).toHaveCount(1);
  await expect(page.getByTestId("version-active-badge")).toBeVisible();

  await page.goto(`/orgs/${seeded.orgId}`);
  await page.getByTestId("api-keys-link").click();
  await page.getByTestId("create-key").click();
  await page.getByLabel("Nome", { exact: true }).fill("Plugin UE5");
  await page.getByRole("button", { name: "Criar" }).click();
  const secret = await page.getByTestId("key-secret").textContent();
  expect(secret).toMatch(/^pol_/);
  await page.getByRole("button", { name: "Fechar" }).click();

  const v1 = await engineHash(seeded.missionId, secret as string);
  expect(v1.status).toBe(200);
  expect(v1.hash).toMatch(/^[0-9a-f]{64}$/);

  await page.goto(seeded.path);
  await canvas.addNode("OBJECTIVE");
  await canvas.dragNodeBy("OBJECTIVE", -60, -60);
  await canvas.connect("DIALOGUE", "OBJECTIVE");
  await canvas.connect("OBJECTIVE", "END");
  await expect(canvas.errorBadges).toHaveCount(0);
  await canvas.save();
  await expect(canvas.dirtyIndicator).toHaveCount(0);

  await page.getByTestId("publish-button").click();
  await page.getByTestId("confirm-publish").click();
  await expect(page.getByTestId("version-item")).toHaveCount(2);

  const v2 = await engineHash(seeded.missionId, secret as string);
  expect(v2.status).toBe(200);
  expect(v2.hash).not.toBe(v1.hash);

  await canvas.dragNodeBy("OBJECTIVE", 40, 40);
  await canvas.save();
  await expect(canvas.dirtyIndicator).toHaveCount(0);
  await page.getByTestId("publish-button").click();
  await page.getByTestId("confirm-publish").click();
  await expect(
    page.getByText("Nada mudou — o conteúdo é idêntico à versão ativa")
  ).toBeVisible();
  await expect(page.getByTestId("version-item")).toHaveCount(2);

  const still = await engineHash(seeded.missionId, secret as string);
  expect(still.hash).toBe(v2.hash);
});

test("revogar a chave corta o engine com 401", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, linearGraph);
  await fetch(
    `${apiUrl}/organizations/${seeded.orgId}/workspaces/${seeded.workspaceId}/missions/${seeded.missionId}/publish`,
    {
      method: "POST",
      headers: { cookie: `session=${sessionToken}` },
    }
  );
  const createRes = await fetch(
    `${apiUrl}/organizations/${seeded.orgId}/api-keys`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        cookie: `session=${sessionToken}`,
      },
      body: JSON.stringify({ name: "Chave descartável" }),
    }
  );
  const created = (await createRes.json()) as { key: string };

  expect((await engineHash(seeded.missionId, created.key)).status).toBe(200);

  await page.goto(`/orgs/${seeded.orgId}/api-keys`);
  await page.getByLabel("Revogar Chave descartável").click();
  await page.getByRole("button", { name: "Revogar" }).click();
  await expect(page.getByTestId("key-revoked")).toBeVisible();

  expect((await engineHash(seeded.missionId, created.key)).status).toBe(401);
});
