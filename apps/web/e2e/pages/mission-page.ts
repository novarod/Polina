import type { Locator, Page } from "@playwright/test";

export class MissionPage {
  readonly editorCanvas: Locator;
  readonly statusBadge: Locator;
  readonly breadcrumb: Locator;

  constructor(readonly page: Page) {
    this.editorCanvas = page.getByTestId("mission-editor");
    this.statusBadge = page.getByTestId("mission-status");
    this.breadcrumb = page.getByTestId("breadcrumb");
  }
}
