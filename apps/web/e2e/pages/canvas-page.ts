import type { Locator, Page } from "@playwright/test";

export class CanvasPage {
  readonly canvas: Locator;
  readonly editor: Locator;
  readonly emptyState: Locator;
  readonly questNodes: Locator;
  readonly edges: Locator;
  readonly minimap: Locator;
  readonly panel: Locator;
  readonly panelData: Locator;
  readonly dataEditor: Locator;
  readonly dirtyIndicator: Locator;
  readonly errorCount: Locator;
  readonly errorBadges: Locator;
  readonly apiErrorBanner: Locator;
  readonly saveButton: Locator;

  constructor(readonly page: Page) {
    this.canvas = page.getByTestId("mission-canvas");
    this.editor = page.getByTestId("mission-editor");
    this.emptyState = page.getByTestId("canvas-empty");
    this.questNodes = page.getByTestId("quest-node");
    this.edges = page.locator(".react-flow__edge");
    this.minimap = page.locator(".react-flow__minimap");
    this.panel = page.getByTestId("node-panel");
    this.panelData = page.getByTestId("node-panel-data");
    this.dataEditor = page.getByTestId("node-data-editor");
    this.dirtyIndicator = page.getByTestId("dirty-indicator");
    this.errorCount = page.getByTestId("graph-error-count");
    this.errorBadges = page.getByTestId("node-error-badge");
    this.apiErrorBanner = page.getByTestId("graph-api-error");
    this.saveButton = page.getByRole("button", { name: "Salvar" });
  }

  nodeByType(type: string): Locator {
    return this.page.locator(
      `[data-testid="quest-node"][data-node-type="${type}"]`
    );
  }

  async openNode(type: string): Promise<void> {
    await this.nodeByType(type).click();
  }

  async addNode(type: string): Promise<void> {
    await this.page.getByTestId(`palette-${type}`).click();
  }

  async dragNodeBy(type: string, dx: number, dy: number): Promise<void> {
    const box = await this.nodeByType(type).boundingBox();
    if (!box) {
      throw new Error(`nó ${type} sem bounding box`);
    }
    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;
    await this.page.mouse.move(startX, startY);
    await this.page.mouse.down();
    await this.page.mouse.move(startX + dx, startY + dy, { steps: 10 });
    await this.page.mouse.up();
  }

  async connect(fromType: string, toType: string): Promise<void> {
    const source = this.nodeByType(fromType).locator(
      ".react-flow__handle-bottom"
    );
    const target = this.nodeByType(toType).locator(".react-flow__handle-top");
    const sourceBox = await source.boundingBox();
    const targetBox = await target.boundingBox();
    if (!sourceBox || !targetBox) {
      throw new Error("handles sem bounding box");
    }
    const startX = sourceBox.x + sourceBox.width / 2;
    const startY = sourceBox.y + sourceBox.height / 2;
    await this.page.mouse.move(startX, startY);
    await this.page.mouse.down();
    await this.page.mouse.move(startX + 4, startY + 4, { steps: 2 });
    await this.page.mouse.move(
      targetBox.x + targetBox.width / 2,
      targetBox.y + targetBox.height / 2,
      { steps: 12 }
    );
    await this.page.waitForTimeout(150);
    await this.page.mouse.up();
  }

  async deleteEdgeBetween(fromType: string, toType: string): Promise<void> {
    const source = await this.nodeByType(fromType)
      .locator(".react-flow__handle-bottom")
      .boundingBox();
    const target = await this.nodeByType(toType)
      .locator(".react-flow__handle-top")
      .boundingBox();
    if (!source || !target) {
      throw new Error("handles sem bounding box");
    }
    const midX =
      (source.x + source.width / 2 + target.x + target.width / 2) / 2;
    const midY =
      (source.y + source.height / 2 + target.y + target.height / 2) / 2;
    await this.page.mouse.click(midX, midY);
    await this.page.keyboard.press("Delete");
  }

  async applyNodeData(json: string): Promise<void> {
    await this.dataEditor.fill(json);
    await this.page.getByRole("button", { name: "Aplicar" }).click();
  }

  async closePanel(): Promise<void> {
    await this.page.getByLabel("Fechar painel").click();
  }

  async save(): Promise<void> {
    await this.saveButton.click();
  }
}
