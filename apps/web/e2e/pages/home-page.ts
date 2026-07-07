import type { Locator, Page } from "@playwright/test";

export class HomePage {
  readonly greeting: Locator;
  readonly logoutButton: Locator;

  constructor(readonly page: Page) {
    this.greeting = page.getByTestId("session-user");
    this.logoutButton = page.getByRole("button", { name: "Sair" });
  }

  async goto(): Promise<void> {
    await this.page.goto("/home");
  }

  async logout(): Promise<void> {
    await this.logoutButton.click();
  }
}
