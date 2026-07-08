import { expect, test } from "../fixtures/session";
import { seedMission } from "../fixtures/seed";
import { CanvasPage } from "../pages/canvas-page";

const branchingGraph = {
  nodes: [
    { id: "inicio", type: "START" },
    { id: "conversa", type: "DIALOGUE", data: { npc: "Aldeão", falas: 3 } },
    { id: "cacada", type: "KILL", data: { alvo: "Lobo", quantidade: 5 } },
    { id: "coleta", type: "COLLECT", data: { item: "Erva rara" } },
    { id: "fim", type: "END" },
  ],
  edges: [
    { id: "e1", source: "inicio", target: "conversa" },
    { id: "e2", source: "conversa", target: "cacada" },
    { id: "e3", source: "conversa", target: "coleta" },
    { id: "e4", source: "cacada", target: "fim" },
    { id: "e5", source: "coleta", target: "fim" },
  ],
};

test.beforeEach(async ({ context, sessionToken }) => {
  await context.addCookies([
    {
      name: "session",
      value: sessionToken,
      url: "http://localhost:3000",
      httpOnly: true,
      sameSite: "Strict",
    },
  ]);
});

test("mission com grafo mostra os nós custom, minimap e painel com data", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, branchingGraph);
  const canvas = new CanvasPage(page);

  await page.goto(seeded.path);

  await expect(canvas.canvas).toBeVisible();
  await expect(canvas.questNodes).toHaveCount(5);
  for (const type of ["START", "DIALOGUE", "KILL", "COLLECT", "END"]) {
    await expect(canvas.nodeByType(type)).toHaveCount(1);
  }
  await expect(canvas.minimap).toBeVisible();

  await canvas.openNode("DIALOGUE");
  await expect(canvas.panel).toBeVisible();
  await expect(canvas.panelData).toContainText("Aldeão");
});

test("mission com grafo vazio mostra o empty state do canvas", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, null);
  const canvas = new CanvasPage(page);

  await page.goto(seeded.path);

  await expect(canvas.emptyState).toBeVisible();
  await expect(canvas.canvas).toHaveCount(0);
});
