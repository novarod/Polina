import type { Locator, Page } from "@playwright/test";

export class WorkspacePage {
  readonly heading: Locator;
  readonly breadcrumb: Locator;

  constructor(readonly page: Page) {
    this.heading = page.getByRole("heading", { name: "Missões" });
    this.breadcrumb = page.getByTestId("breadcrumb");
  }

  missionCard(name: string): Locator {
    return this.page.getByTestId("mission-card").filter({ hasText: name });
  }

  async createMission(name: string, description: string): Promise<void> {
    await this.page.getByTestId("create-mission").click();
    await this.page.getByLabel("Nome", { exact: true }).fill(name);
    await this.page.getByLabel("Descrição (opcional)").fill(description);
    await this.page.getByRole("button", { name: "Criar" }).click();
  }

  async openMission(name: string): Promise<void> {
    await this.missionCard(name).getByRole("link").click();
  }

  async renameMission(currentName: string, newName: string): Promise<void> {
    await this.page.getByLabel(`Editar ${currentName}`).click();
    await this.page.getByLabel("Nome", { exact: true }).fill(newName);
    await this.page.getByRole("button", { name: "Salvar" }).click();
  }

  async deleteMission(name: string): Promise<void> {
    await this.page.getByLabel(`Deletar ${name}`).click();
    await this.page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Deletar" })
      .click();
  }
}
