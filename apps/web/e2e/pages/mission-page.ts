import type { Locator, Page } from "@playwright/test";

export class MissionPage {
  readonly emptyCanvas: Locator;
  readonly statusBadge: Locator;
  readonly breadcrumb: Locator;

  constructor(readonly page: Page) {
    this.emptyCanvas = page.getByTestId("canvas-empty");
    this.statusBadge = page.getByTestId("mission-status");
    this.breadcrumb = page.getByTestId("breadcrumb");
  }
}
