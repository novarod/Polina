import { expect, test } from "../fixtures/session";
import { HomePage } from "../pages/home-page";
import { LoginPage } from "../pages/login-page";

test("login com credenciais válidas leva à home com o nome do usuário", async ({
  page,
  account,
}) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(account.email, account.password);

  const homePage = new HomePage(page);
  await expect(page).toHaveURL("/home");
  await expect(homePage.greeting).toContainText(account.name);
});

test("credenciais inválidas mostram erro e permanecem em /login", async ({
  page,
  account,
}) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(account.email, "senha-errada-123");

  await expect(loginPage.errorAlert).toHaveText("Email ou senha inválidos");
  await expect(page).toHaveURL("/login");
});

test("acessar /home sem sessão redireciona para /login", async ({ page }) => {
  const homePage = new HomePage(page);
  await homePage.goto();

  await expect(page).toHaveURL("/login");
});

test("reload em /home mantém a sessão", async ({ page, account }) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(account.email, account.password);
  await expect(page).toHaveURL("/home");

  await page.reload();

  const homePage = new HomePage(page);
  await expect(page).toHaveURL("/home");
  await expect(homePage.greeting).toContainText(account.name);
});

test("logout encerra a sessão e /home volta a redirecionar", async ({
  page,
  account,
}) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(account.email, account.password);
  await expect(page).toHaveURL("/home");

  const homePage = new HomePage(page);
  await homePage.logout();
  await expect(page).toHaveURL("/login");

  await homePage.goto();
  await expect(page).toHaveURL("/login");
});
