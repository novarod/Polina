import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CreateOrgDialog } from "@/components/orgs/create-org-dialog";
import { ApiError } from "@/services/api";
import { createOrganization } from "@/services/organizations";

vi.mock("@/services/organizations");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess },
}));

async function openDialog() {
  const user = userEvent.setup();
  render(<CreateOrgDialog />);
  await user.click(screen.getByTestId("create-org"));
  return user;
}

describe("CreateOrgDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("validates name and slug before calling the service", async () => {
    const user = await openDialog();

    await user.type(screen.getByLabelText("Slug"), "Meu Estúdio");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    expect(
      await screen.findByText("O nome precisa de pelo menos 2 caracteres")
    ).toBeVisible();
    expect(
      screen.getByText(
        "Use apenas letras minúsculas, números e hífens (ex.: meu-estudio)"
      )
    ).toBeVisible();
    expect(createOrganization).not.toHaveBeenCalled();
  });

  it("creates the organization, closes the dialog and refreshes", async () => {
    vi.mocked(createOrganization).mockResolvedValue({
      id: "o1",
      name: "Acme",
      slug: "acme",
      created_at: "2026-07-07T00:00:00Z",
    });
    const user = await openDialog();

    await user.type(screen.getByLabelText("Nome"), "Acme");
    await user.type(screen.getByLabelText("Slug"), "acme");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    await waitFor(() => {
      expect(createOrganization).toHaveBeenCalledWith({
        name: "Acme",
        slug: "acme",
      });
      expect(refresh).toHaveBeenCalled();
      expect(toastSuccess).toHaveBeenCalledWith("Organização criada");
    });
    expect(screen.queryByLabelText("Slug")).not.toBeInTheDocument();
  });

  it("shows the API error inside the dialog and keeps it open", async () => {
    vi.mocked(createOrganization).mockRejectedValue(
      new ApiError(422, "slug already in use")
    );
    const user = await openDialog();

    await user.type(screen.getByLabelText("Nome"), "Acme");
    await user.type(screen.getByLabelText("Slug"), "acme");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    expect(await screen.findByTestId("dialog-error")).toHaveTextContent(
      "slug already in use"
    );
    expect(screen.getByLabelText("Slug")).toHaveValue("acme");
    expect(refresh).not.toHaveBeenCalled();
  });
});
