import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LoginForm } from "@/components/auth/login-form";
import { ApiError } from "@/services/api";
import { login } from "@/services/auth";

vi.mock("@/services/auth");

const push = vi.fn();
const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, refresh }),
}));

describe("LoginForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows zod validation messages and does not call the service", async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByText("Informe um email válido")).toBeVisible();
    expect(screen.getByText("Informe a senha")).toBeVisible();
    expect(login).not.toHaveBeenCalled();
  });

  it("submits credentials and navigates to /home on success", async () => {
    vi.mocked(login).mockResolvedValue({ user_id: "u1", name: "Alice" });
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.type(screen.getByLabelText("Email"), "a@b.com");
    await user.type(screen.getByLabelText("Senha"), "hunter22");
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith({
        email: "a@b.com",
        password: "hunter22",
      });
      expect(push).toHaveBeenCalledWith("/home");
    });
  });

  it("shows a credentials error on 401 and stays on the page", async () => {
    vi.mocked(login).mockRejectedValue(new ApiError(401, "invalid credentials"));
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.type(screen.getByLabelText("Email"), "a@b.com");
    await user.type(screen.getByLabelText("Senha"), "wrong");
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Email ou senha inválidos"
    );
    expect(push).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Email")).toHaveValue("a@b.com");
  });

  it("shows the network error message when the API is unreachable", async () => {
    vi.mocked(login).mockRejectedValue(
      new ApiError(0, "Não foi possível conectar ao servidor")
    );
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.type(screen.getByLabelText("Email"), "a@b.com");
    await user.type(screen.getByLabelText("Senha"), "hunter22");
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Não foi possível conectar ao servidor"
    );
  });
});
