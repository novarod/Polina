import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useOrgStatus } from "@/hooks/use-org-status";
import type { RealtimeServerFrame } from "@/types/realtime";

const client = vi.hoisted(() => ({
  subscribeMissionPos: vi.fn(),
  subscribeOrgStatus: vi.fn(),
  onSelf: vi.fn(),
  getSelf: vi.fn(),
  sendPos: vi.fn(),
  setEditing: vi.fn(),
}));

vi.mock("@/services/realtime", () => ({ realtimeClient: client }));

let statusListener: (frame: RealtimeServerFrame) => void;
const unsubscribe = vi.fn();

describe("useOrgStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    client.subscribeOrgStatus.mockImplementation(
      (_orgId, listener: (frame: RealtimeServerFrame) => void) => {
        statusListener = listener;
        return unsubscribe;
      }
    );
  });

  it("snapshot substitui contagens e ignora missões zeradas", () => {
    const { result } = renderHook(() => useOrgStatus("o1"));
    act(() => {
      statusListener({
        type: "snapshot",
        topic: "status:org:o1",
        missions: [
          { mission_id: "m1", count: 2 },
          { mission_id: "m2", count: 0 },
        ],
      });
    });
    expect(result.current).toEqual({ m1: 2 });
  });

  it("eventos status atualizam e removem contagens", () => {
    const { result } = renderHook(() => useOrgStatus("o1"));
    act(() => {
      statusListener({
        type: "status",
        topic: "status:org:o1",
        mission_id: "m1",
        count: 1,
      });
    });
    expect(result.current).toEqual({ m1: 1 });

    act(() => {
      statusListener({
        type: "status",
        topic: "status:org:o1",
        mission_id: "m1",
        count: 0,
      });
    });
    expect(result.current).toEqual({});
  });

  it("cancela assinatura ao desmontar", () => {
    const { unmount } = renderHook(() => useOrgStatus("o1"));
    unmount();
    expect(unsubscribe).toHaveBeenCalled();
  });
});
