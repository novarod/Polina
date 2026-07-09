import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { VersionList } from "@/components/missions/version-list";
import { getMissionVersion } from "@/services/missions";
import type { MissionVersion } from "@/types/mission";

vi.mock("@/services/missions");

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess, error: vi.fn() },
}));

const versions: MissionVersion[] = [
  {
    id: "v2",
    version_number: 2,
    hash: "b".repeat(64),
    published_by_id: "u1",
    created_at: "2026-07-08T12:00:00Z",
  },
  {
    id: "v1",
    version_number: 1,
    hash: "a".repeat(64),
    published_by_id: "u1",
    created_at: "2026-07-07T12:00:00Z",
  },
];

function renderList(activeHash: string | null) {
  render(
    <VersionList
      orgId="o1"
      workspaceId="w1"
      missionId="m1"
      versions={versions}
      activeHash={activeHash}
    />
  );
}

describe("VersionList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows the empty state when nothing was published", () => {
    render(
      <VersionList
        orgId="o1"
        workspaceId="w1"
        missionId="m1"
        versions={[]}
        activeHash={null}
      />
    );
    expect(
      screen.getByText("Nenhuma versão publicada ainda.")
    ).toBeVisible();
  });

  it("marks only the active version", () => {
    renderList("b".repeat(64));

    const items = screen.getAllByTestId("version-item");
    expect(items).toHaveLength(2);
    expect(screen.getAllByTestId("version-active-badge")).toHaveLength(1);
    expect(items[0]).toHaveTextContent("v2");
    expect(items[0]).toHaveTextContent("Ativa");
  });

  it("copies the full hash to the clipboard", async () => {
    const user = userEvent.setup();
    const writeText = vi
      .spyOn(navigator.clipboard, "writeText")
      .mockResolvedValue();
    renderList(null);

    await user.click(screen.getByLabelText("Copiar hash da v2"));

    expect(writeText).toHaveBeenCalledWith("b".repeat(64));
    expect(toastSuccess).toHaveBeenCalledWith("Hash copiado");
  });

  it("loads and shows the compiled contract on expand", async () => {
    vi.mocked(getMissionVersion).mockResolvedValue({
      id: "v1",
      version_number: 1,
      hash: "a".repeat(64),
      published_by_id: "u1",
      created_at: "2026-07-07T12:00:00Z",
      mission_data: { mission_id: "m1", start_node: "inicio" },
    });
    const user = userEvent.setup();
    renderList(null);

    await user.click(screen.getByRole("button", { name: /^v1\b/ }));

    await waitFor(() => {
      expect(getMissionVersion).toHaveBeenCalledWith(
        "o1",
        "w1",
        "m1",
        "a".repeat(64)
      );
    });
    expect(await screen.findByTestId("version-data")).toHaveTextContent(
      "start_node"
    );
  });
});
