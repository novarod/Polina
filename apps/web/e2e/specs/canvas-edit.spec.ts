import { expect, test } from "../fixtures/session";
import { seedMission } from "../fixtures/seed";
import { CanvasPage } from "../pages/canvas-page";

const linearGraph = {
  nodes: [
    { id: "inicio", type: "START", position: { x: 0, y: 0 } },
    {
      id: "conversa",
      type: "DIALOGUE",
      data: { npc: "Aldeão" },
      position: { x: 0, y: 180 },
    },
    { id: "fim", type: "END", position: { x: 0, y: 360 } },
  ],
  edges: [
    { id: "e1", source: "inicio", target: "conversa" },
    { id: "e2", source: "conversa", target: "fim" },
  ],
};

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

test("designer monta uma quest do zero, salva e recarrega persistido", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, null);
  const canvas = new CanvasPage(page);
  await page.goto(seeded.path);

  await expect(page.getByTestId("node-palette")).toBeVisible();

  await canvas.addNode("START");
  await canvas.dragNodeBy("START", -140, -150);
  await canvas.addNode("DIALOGUE");
  await canvas.dragNodeBy("DIALOGUE", 120, -40);
  await canvas.addNode("END");
  await canvas.dragNodeBy("END", -40, 130);
  await expect(canvas.questNodes).toHaveCount(3);

  await canvas.connect("START", "DIALOGUE");
  await canvas.connect("DIALOGUE", "END");
  await expect(canvas.edges).toHaveCount(2);

  await canvas.openNode("DIALOGUE");
  await canvas.applyNodeData('{"npc": "Aldeão", "falas": 3}');
  await canvas.closePanel();

  await expect(canvas.dirtyIndicator).toBeVisible();
  await expect(canvas.errorCount).toHaveCount(0);
  await canvas.save();
  await expect(canvas.dirtyIndicator).toHaveCount(0);

  await page.reload();
  await expect(canvas.questNodes).toHaveCount(3);
  await expect(canvas.edges).toHaveCount(2);

  const startBox = await canvas.nodeByType("START").boundingBox();
  const dialogueBox = await canvas.nodeByType("DIALOGUE").boundingBox();
  const endBox = await canvas.nodeByType("END").boundingBox();
  expect(startBox && dialogueBox && startBox.y < dialogueBox.y).toBe(true);
  expect(dialogueBox && endBox && dialogueBox.y < endBox.y).toBe(true);

  await canvas.openNode("DIALOGUE");
  await expect(canvas.dataEditor).toHaveValue(/Aldeão/);
});

test("dead-end aparece ao vivo, 422 real no save e o fix limpa tudo", async ({
  page,
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, linearGraph);
  const canvas = new CanvasPage(page);
  await page.goto(seeded.path);
  await expect(canvas.questNodes).toHaveCount(3);

  await canvas.deleteEdgeBetween("DIALOGUE", "END");
  await expect(canvas.edges).toHaveCount(1);
  await expect(
    canvas.nodeByType("DIALOGUE").getByTestId("node-error-badge")
  ).toBeVisible();
  await expect(canvas.errorCount).toBeVisible();

  await canvas.save();
  await expect(canvas.apiErrorBanner).toBeVisible();
  await expect(canvas.apiErrorBanner).toContainText("dag validation failed");

  await canvas.connect("DIALOGUE", "END");
  await expect(canvas.errorBadges).toHaveCount(0);
  await canvas.save();
  await expect(canvas.dirtyIndicator).toHaveCount(0);
  await expect(canvas.apiErrorBanner).toHaveCount(0);
});
