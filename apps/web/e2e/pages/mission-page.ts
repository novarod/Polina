import type { Locator, Page } from "@playwright/test";

export class MissionPage {
  readonly canvasPlaceholder: Locator;
  readonly statusBadge: Locator;
  readonly breadcrumb: Locator;

  constructor(readonly page: Page) {
    this.canvasPlaceholder = page.getByTestId("canvas-placeholder");
    this.statusBadge = page.getByTestId("mission-status");
    this.breadcrumb = page.getByTestId("breadcrumb");
  }
}
