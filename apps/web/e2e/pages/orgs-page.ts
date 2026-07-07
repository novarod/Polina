import type { Locator, Page } from "@playwright/test";

export class OrgsPage {
  readonly heading: Locator;
  readonly userMenu: Locator;
  readonly createButton: Locator;

  constructor(readonly page: Page) {
    this.heading = page.getByRole("heading", { name: "Organizações" });
    this.userMenu = page.getByTestId("user-menu");
    this.createButton = page.getByTestId("create-org");
  }

  async goto(): Promise<void> {
    await this.page.goto("/orgs");
  }

  orgCard(name: string): Locator {
    return this.page.getByTestId("org-card").filter({ hasText: name });
  }

  async createOrg(name: string, slug: string): Promise<void> {
    await this.createButton.click();
    await this.page.getByLabel("Nome", { exact: true }).fill(name);
    await this.page.getByLabel("Slug").fill(slug);
    await this.page.getByRole("button", { name: "Criar" }).click();
  }

  async openOrg(name: string): Promise<void> {
    await this.orgCard(name).getByRole("link").click();
  }

  async deleteOrg(name: string): Promise<void> {
    await this.page.getByLabel(`Deletar ${name}`).click();
    await this.page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Deletar" })
      .click();
  }

  async logout(): Promise<void> {
    await this.userMenu.click();
    await this.page.getByRole("menuitem", { name: "Sair" }).click();
  }
}
