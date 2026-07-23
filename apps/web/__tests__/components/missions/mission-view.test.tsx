import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { MissionView } from "@/components/missions/mission-view";
import type { Mission } from "@/types/mission";

vi.mock("@/services/missions");

vi.mock("@/services/realtime", () => ({
  realtimeClient: {
    subscribeMissionPos: vi.fn(() => () => {}),
    subscribeOrgStatus: vi.fn(() => () => {}),
    onSelf: vi.fn(() => () => {}),
    sendPos: vi.fn(),
    setEditing: vi.fn(),
  },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: vi.fn() }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), info: vi.fn(), error: vi.fn() },
}));

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

beforeAll(() => {
  globalThis.ResizeObserver =
    globalThis.ResizeObserver ?? ResizeObserverStub;
});

const mission: Mission = {
  id: "m1",
  organization_id: "o1",
  workspace_id: "w1",
  name: "Resgate",
  description: "",
  status: "DRAFT",
  active_hash: null,
  graph: {
    nodes: [
      { id: "start-1", type: "START", position: { x: 0, y: 0 } },
      { id: "end-1", type: "END", position: { x: 0, y: 200 } },
    ],
    edges: [{ id: "e1", source: "start-1", target: "end-1" }],
  },
  created_by_id: "u1",
  created_at: "2026-07-08T12:00:00Z",
};

describe("MissionView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("hides the publish button below DESIGNER", () => {
    render(
      <MissionView
        mission={mission}
        versions={[]}
        canEdit={false}
        orgId="o1"
        workspaceId="w1"
      />
    );

    expect(screen.queryByTestId("publish-button")).not.toBeInTheDocument();
    expect(screen.queryByTestId("node-palette")).not.toBeInTheDocument();
  });

  it("disables publish while the editor is dirty", async () => {
    const user = userEvent.setup();
    render(
      <MissionView
        mission={mission}
        versions={[]}
        canEdit
        orgId="o1"
        workspaceId="w1"
      />
    );

    expect(screen.getByTestId("publish-button")).toBeEnabled();

    await user.click(screen.getByTestId("palette-OBJECTIVE"));

    expect(screen.getByTestId("publish-button")).toBeDisabled();
  });
});
