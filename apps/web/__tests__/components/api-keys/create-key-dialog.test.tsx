import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CreateKeyDialog } from "@/components/api-keys/create-key-dialog";
import { createApiKey } from "@/services/api-keys";

vi.mock("@/services/api-keys");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess, error: vi.fn() },
}));

describe("CreateKeyDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("creates the key and shows the secret exactly once", async () => {
    vi.mocked(createApiKey).mockResolvedValue({
      id: "k1",
      name: "Plugin UE5",
      key: "pol_secreto123",
      created_at: "2026-07-08T12:00:00Z",
    });
    const user = userEvent.setup();
    const writeText = vi
      .spyOn(navigator.clipboard, "writeText")
      .mockResolvedValue();
    render(<CreateKeyDialog orgId="o1" />);

    await user.click(screen.getByTestId("create-key"));
    await user.type(screen.getByLabelText("Nome"), "Plugin UE5");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    await waitFor(() => {
      expect(createApiKey).toHaveBeenCalledWith("o1", "Plugin UE5");
    });
    expect(await screen.findByTestId("key-secret")).toHaveTextContent(
      "pol_secreto123"
    );
    expect(refresh).toHaveBeenCalled();

    await user.click(screen.getByLabelText("Copiar chave"));
    expect(writeText).toHaveBeenCalledWith("pol_secreto123");

    await user.click(screen.getByRole("button", { name: "Fechar" }));
    await waitFor(() => {
      expect(screen.queryByTestId("key-secret")).not.toBeInTheDocument();
    });
  });
});
