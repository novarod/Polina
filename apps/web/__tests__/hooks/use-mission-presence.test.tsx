import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useMissionPresence } from "@/hooks/use-mission-presence";
import type { PresenceUser, RealtimeServerFrame } from "@/types/realtime";

const client = vi.hoisted(() => ({
  subscribeMissionPos: vi.fn(),
  subscribeOrgStatus: vi.fn(),
  onSelf: vi.fn(),
  getSelf: vi.fn(),
  sendPos: vi.fn(),
  setEditing: vi.fn(),
}));

vi.mock("@/services/realtime", () => ({ realtimeClient: client }));

let missionListener: (frame: RealtimeServerFrame) => void;
let selfListener: (user: PresenceUser) => void;
const unsubscribeMission = vi.fn();

const me = { id: "me", name: "Eu Mesmo" };
const peer = { id: "peer", name: "Outra Pessoa" };
const options = {
  orgId: "o1",
  workspaceId: "w1",
  missionId: "m1",
  editing: true,
};

describe("useMissionPresence", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    client.subscribeMissionPos.mockImplementation(
      (_params, listener: (frame: RealtimeServerFrame) => void) => {
        missionListener = listener;
        return unsubscribeMission;
      }
    );
    client.onSelf.mockImplementation((listener: (user: PresenceUser) => void) => {
      selfListener = listener;
      return () => {};
    });
  });

  it("snapshot substitui usuários e peers exclui a própria identidade", () => {
    const { result } = renderHook(() => useMissionPresence(options));

    act(() => {
      selfListener(me);
      missionListener({ type: "snapshot", topic: "t", users: [me, peer] });
    });
    expect(result.current.users).toHaveLength(2);
    expect(result.current.peers).toEqual([peer]);
    expect(result.current.self).toEqual(me);

    act(() => {
      missionListener({ type: "snapshot", topic: "t", users: [me] });
    });
    expect(result.current.peers).toEqual([]);
  });

  it("join e leave atualizam presença sem duplicar", () => {
    const { result } = renderHook(() => useMissionPresence(options));

    act(() => {
      selfListener(me);
      missionListener({ type: "snapshot", topic: "t", users: [me] });
      missionListener({ type: "join", topic: "t", user: peer });
      missionListener({ type: "join", topic: "t", user: peer });
    });
    expect(result.current.peers).toEqual([peer]);

    act(() => {
      missionListener({ type: "leave", topic: "t", user: peer });
    });
    expect(result.current.peers).toEqual([]);
  });

  it("repassa cursores de peers e filtra o cursor próprio", () => {
    const { result } = renderHook(() => useMissionPresence(options));
    const moves: Array<{ userId: string; x: number; y: number }> = [];
    act(() => {
      selfListener(me);
    });
    result.current.subscribeCursor((move) => moves.push(move));

    act(() => {
      missionListener({ type: "pos", topic: "t", user_id: "me", x: 1, y: 1 });
      missionListener({ type: "pos", topic: "t", user_id: "peer", x: 5, y: 9 });
    });
    expect(moves).toEqual([{ userId: "peer", x: 5, y: 9 }]);
  });

  it("declara editando ao montar e remove ao desmontar", () => {
    const { unmount } = renderHook(() => useMissionPresence(options));
    expect(client.setEditing).toHaveBeenCalledWith(
      { orgId: "o1", workspaceId: "w1", missionId: "m1" },
      true
    );
    unmount();
    expect(client.setEditing).toHaveBeenLastCalledWith(
      { orgId: "o1", workspaceId: "w1", missionId: "m1" },
      false
    );
    expect(unsubscribeMission).toHaveBeenCalled();
  });

  it("não declara editando para viewers", () => {
    renderHook(() => useMissionPresence({ ...options, editing: false }));
    expect(client.setEditing).not.toHaveBeenCalled();
  });

  it("sendPos delega ao singleton com o mission id", () => {
    const { result } = renderHook(() => useMissionPresence(options));
    result.current.sendPos(3, 4);
    expect(client.sendPos).toHaveBeenCalledWith("m1", 3, 4);
  });
});
