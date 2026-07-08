import type { Locator, Page } from "@playwright/test";

export class CanvasPage {
  readonly canvas: Locator;
  readonly emptyState: Locator;
  readonly questNodes: Locator;
  readonly minimap: Locator;
  readonly panel: Locator;
  readonly panelData: Locator;

  constructor(readonly page: Page) {
    this.canvas = page.getByTestId("mission-canvas");
    this.emptyState = page.getByTestId("canvas-empty");
    this.questNodes = page.getByTestId("quest-node");
    this.minimap = page.locator(".react-flow__minimap");
    this.panel = page.getByTestId("node-panel");
    this.panelData = page.getByTestId("node-panel-data");
  }

  nodeByType(type: string): Locator {
    return this.page.locator(
      `[data-testid="quest-node"][data-node-type="${type}"]`
    );
  }

  async openNode(type: string): Promise<void> {
    await this.nodeByType(type).click();
  }
}
