import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { PublishButton } from "@/components/missions/publish-button";
import { ApiError } from "@/services/api";
import { publishMission } from "@/services/missions";

vi.mock("@/services/missions");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

const toastSuccess = vi.hoisted(() => vi.fn());
const toastInfo = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess, info: toastInfo },
}));

function renderButton(dirty: boolean, activeHash: string | null = null) {
  render(
    <PublishButton
      orgId="o1"
      workspaceId="w1"
      missionId="m1"
      activeHash={activeHash}
      dirty={dirty}
    />
  );
}

describe("PublishButton", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is disabled with a hint while the canvas is dirty", () => {
    renderButton(true);
    const button = screen.getByTestId("publish-button");
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("title", "Salve o grafo antes de publicar");
  });

  it("publishes a new version after confirmation", async () => {
    vi.mocked(publishMission).mockResolvedValue({
      mission_id: "m1",
      version: 3,
      hash: "novo-hash",
      status: "APPROVED",
      active_hash: "novo-hash",
    });
    const user = userEvent.setup();
    renderButton(false, "hash-antigo");

    await user.click(screen.getByTestId("publish-button"));
    await user.click(screen.getByTestId("confirm-publish"));

    await waitFor(() => {
      expect(publishMission).toHaveBeenCalledWith("o1", "w1", "m1");
      expect(toastSuccess).toHaveBeenCalledWith("Versão v3 publicada");
      expect(refresh).toHaveBeenCalled();
    });
  });

  it("announces when nothing changed (idempotent hash)", async () => {
    vi.mocked(publishMission).mockResolvedValue({
      mission_id: "m1",
      version: 2,
      hash: "mesmo-hash",
      status: "APPROVED",
      active_hash: "mesmo-hash",
    });
    const user = userEvent.setup();
    renderButton(false, "mesmo-hash");

    await user.click(screen.getByTestId("publish-button"));
    await user.click(screen.getByTestId("confirm-publish"));

    await waitFor(() => {
      expect(toastInfo).toHaveBeenCalledWith(
        "Nada mudou — o conteúdo é idêntico à versão ativa"
      );
    });
    expect(toastSuccess).not.toHaveBeenCalled();
  });

  it("keeps the dialog open showing the 422 from the API", async () => {
    vi.mocked(publishMission).mockRejectedValue(
      new ApiError(422, "dag validation failed: dead-end")
    );
    const user = userEvent.setup();
    renderButton(false);

    await user.click(screen.getByTestId("publish-button"));
    await user.click(screen.getByTestId("confirm-publish"));

    expect(await screen.findByTestId("publish-error")).toHaveTextContent(
      "dag validation failed: dead-end"
    );
    expect(refresh).not.toHaveBeenCalled();
  });
});
