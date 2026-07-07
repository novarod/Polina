import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { RenameMissionDialog } from "@/components/missions/rename-mission-dialog";
import { Button } from "@/components/ui/button";
import { updateMission } from "@/services/missions";
import type { Mission } from "@/types/mission";

vi.mock("@/services/missions");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess },
}));

const mission: Mission = {
  id: "m1",
  organization_id: "o1",
  workspace_id: "w1",
  name: "Resgate",
  description: "Missão de resgate",
  status: "DRAFT",
  active_hash: null,
  graph: { nodes: [], edges: [] },
  created_by_id: "u1",
  created_at: "2026-07-07T00:00:00Z",
};

describe("RenameMissionDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("prefills current values and submits the update", async () => {
    vi.mocked(updateMission).mockResolvedValue({
      ...mission,
      name: "Resgate na Vila",
    });
    const user = userEvent.setup();
    render(
      <RenameMissionDialog
        trigger={<Button>Editar</Button>}
        mission={mission}
      />
    );
    await user.click(screen.getByRole("button", { name: "Editar" }));

    const nameInput = screen.getByLabelText("Nome");
    expect(nameInput).toHaveValue("Resgate");

    await user.clear(nameInput);
    await user.type(nameInput, "Resgate na Vila");
    await user.click(screen.getByRole("button", { name: "Salvar" }));

    await waitFor(() => {
      expect(updateMission).toHaveBeenCalledWith("o1", "w1", "m1", {
        name: "Resgate na Vila",
        description: "Missão de resgate",
      });
      expect(toastSuccess).toHaveBeenCalledWith("Missão atualizada");
      expect(refresh).toHaveBeenCalled();
    });
  });
});
