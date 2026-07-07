import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CreateWorkspaceDialog } from "@/components/workspaces/create-workspace-dialog";
import { createWorkspace } from "@/services/workspaces";

vi.mock("@/services/workspaces");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess },
}));

describe("CreateWorkspaceDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("validates the name before calling the service", async () => {
    const user = userEvent.setup();
    render(<CreateWorkspaceDialog orgId="o1" />);
    await user.click(screen.getByTestId("create-workspace"));

    await user.click(screen.getByRole("button", { name: "Criar" }));

    expect(
      await screen.findByText("O nome precisa de pelo menos 2 caracteres")
    ).toBeVisible();
    expect(createWorkspace).not.toHaveBeenCalled();
  });

  it("creates the workspace with description and closes", async () => {
    vi.mocked(createWorkspace).mockResolvedValue({
      id: "w1",
      organization_id: "o1",
      name: "Sistemas",
      description: "Combate",
      created_at: "2026-07-07T00:00:00Z",
    });
    const user = userEvent.setup();
    render(<CreateWorkspaceDialog orgId="o1" />);
    await user.click(screen.getByTestId("create-workspace"));

    await user.type(screen.getByLabelText("Nome"), "Sistemas");
    await user.type(
      screen.getByLabelText("Descrição (opcional)"),
      "Combate"
    );
    await user.click(screen.getByRole("button", { name: "Criar" }));

    await waitFor(() => {
      expect(createWorkspace).toHaveBeenCalledWith("o1", {
        name: "Sistemas",
        description: "Combate",
      });
      expect(toastSuccess).toHaveBeenCalledWith("Workspace criado");
    });
    expect(screen.queryByLabelText("Nome")).not.toBeInTheDocument();
  });
});
