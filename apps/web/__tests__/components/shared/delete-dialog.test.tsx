import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { DeleteDialog } from "@/components/shared/delete-dialog";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/services/api";

describe("DeleteDialog", () => {
  const onConfirm = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  async function openDialog() {
    const user = userEvent.setup();
    render(
      <DeleteDialog
        trigger={<Button>Abrir exclusão</Button>}
        entityLabel="a organização"
        name="Acme"
        onConfirm={onConfirm}
      />
    );
    await user.click(screen.getByRole("button", { name: "Abrir exclusão" }));
    return user;
  }

  it("names the target and confirms the deletion", async () => {
    onConfirm.mockResolvedValue(undefined);
    const user = await openDialog();

    expect(
      screen.getByText("Deletar a organização “Acme”?")
    ).toBeVisible();

    await user.click(screen.getByRole("button", { name: "Deletar" }));

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalled();
    });
    expect(
      screen.queryByText("Deletar a organização “Acme”?")
    ).not.toBeInTheDocument();
  });

  it("shows the API error and keeps the dialog open", async () => {
    onConfirm.mockRejectedValue(new ApiError(403, "insufficient role"));
    const user = await openDialog();

    await user.click(screen.getByRole("button", { name: "Deletar" }));

    expect(await screen.findByTestId("dialog-error")).toHaveTextContent(
      "insufficient role"
    );
    expect(screen.getByText("Deletar a organização “Acme”?")).toBeVisible();
  });
});
