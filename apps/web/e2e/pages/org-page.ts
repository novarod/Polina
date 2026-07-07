import type { Locator, Page } from "@playwright/test";

export class OrgPage {
  readonly heading: Locator;
  readonly breadcrumb: Locator;

  constructor(readonly page: Page) {
    this.heading = page.getByRole("heading", { name: "Workspaces" });
    this.breadcrumb = page.getByTestId("breadcrumb");
  }

  workspaceCard(name: string): Locator {
    return this.page.getByTestId("workspace-card").filter({ hasText: name });
  }

  async createWorkspace(name: string, description: string): Promise<void> {
    await this.page.getByTestId("create-workspace").click();
    await this.page.getByLabel("Nome", { exact: true }).fill(name);
    await this.page.getByLabel("Descrição (opcional)").fill(description);
    await this.page.getByRole("button", { name: "Criar" }).click();
  }

  async openWorkspace(name: string): Promise<void> {
    await this.workspaceCard(name).getByRole("link").click();
  }

  async deleteWorkspace(name: string): Promise<void> {
    await this.page.getByLabel(`Deletar ${name}`).click();
    await this.page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Deletar" })
      .click();
  }
}
