import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { RealtimeServerFrame } from "@/types/realtime";

const ticketFetch = vi.hoisted(() => vi.fn());
vi.mock("@/services/api", () => ({
  apiFetch: ticketFetch,
}));

class FakeWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;
  static instances: FakeWebSocket[] = [];

  url: string;
  readyState = FakeWebSocket.CONNECTING;
  sent: string[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  send(data: string) {
    this.sent.push(data);
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  }

  open() {
    this.readyState = FakeWebSocket.OPEN;
    this.onopen?.();
  }

  serverClose() {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  }

  message(frame: RealtimeServerFrame) {
    this.onmessage?.({ data: JSON.stringify(frame) });
  }

  sentFrames(): Array<Record<string, unknown>> {
    return this.sent.map((raw) => JSON.parse(raw) as Record<string, unknown>);
  }
}

async function freshClient() {
  vi.resetModules();
  const mod = await import("@/services/realtime");
  return mod.realtimeClient;
}

const params = { orgId: "o1", workspaceId: "w1", missionId: "m1" };

describe("realtimeClient", () => {
  beforeEach(() => {
    FakeWebSocket.instances = [];
    vi.stubGlobal("WebSocket", FakeWebSocket);
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
    ticketFetch.mockReset();
  });

  it("compartilha uma conexão entre assinantes do mesmo tópico", async () => {
    const client = await freshClient();
    const off1 = client.subscribeMissionPos(params, () => {});
    const off2 = client.subscribeMissionPos(params, () => {});

    expect(FakeWebSocket.instances).toHaveLength(1);
    const socket = FakeWebSocket.instances[0];
    socket.open();

    const subscribes = socket
      .sentFrames()
      .filter((frame) => frame.type === "subscribe");
    expect(subscribes).toHaveLength(1);
    expect(subscribes[0]).toMatchObject({
      plane: "pos",
      org_id: "o1",
      workspace_id: "w1",
      mission_id: "m1",
    });

    off1();
    expect(
      socket.sentFrames().filter((frame) => frame.type === "unsubscribe")
    ).toHaveLength(0);
    expect(socket.readyState).toBe(FakeWebSocket.OPEN);

    off2();
    expect(
      socket.sentFrames().filter((frame) => frame.type === "unsubscribe")
    ).toHaveLength(1);
    expect(socket.readyState).toBe(FakeWebSocket.CLOSED);
  });

  it("usa o rewrite /api/realtime/ws quando NEXT_PUBLIC_REALTIME_URL não está setada", async () => {
    const client = await freshClient();
    const off = client.subscribeMissionPos(params, () => {});
    expect(FakeWebSocket.instances[0].url).toContain("/api/realtime/ws");
    off();
  });

  it("entrega frames apenas aos listeners do tópico", async () => {
    const client = await freshClient();
    const posFrames: RealtimeServerFrame[] = [];
    const statusFrames: RealtimeServerFrame[] = [];
    client.subscribeMissionPos(params, (frame) => posFrames.push(frame));
    client.subscribeOrgStatus("o1", (frame) => statusFrames.push(frame));

    const socket = FakeWebSocket.instances[0];
    socket.open();
    socket.message({ type: "snapshot", topic: "pos:mission:m1", users: [] });
    socket.message({ type: "snapshot", topic: "status:org:o1", missions: [] });

    expect(posFrames).toHaveLength(1);
    expect(statusFrames).toHaveLength(1);
  });

  it("notifica identidade própria a partir do hello", async () => {
    const client = await freshClient();
    client.subscribeMissionPos(params, () => {});
    const socket = FakeWebSocket.instances[0];
    socket.open();

    const seen: string[] = [];
    client.onSelf((user) => seen.push(user.id));
    socket.message({ type: "hello", v: 1, user: { id: "me", name: "Eu" } });

    expect(seen).toEqual(["me"]);
    expect(client.getSelf()?.id).toBe("me");

    const late: string[] = [];
    client.onSelf((user) => late.push(user.id));
    expect(late).toEqual(["me"]);
  });

  it("reconecta com backoff e reenvia subscribe e editing", async () => {
    const client = await freshClient();
    client.subscribeMissionPos(params, () => {});
    const first = FakeWebSocket.instances[0];
    first.open();
    client.setEditing(params, true);

    first.serverClose();
    expect(FakeWebSocket.instances).toHaveLength(1);

    vi.advanceTimersByTime(600);
    expect(FakeWebSocket.instances).toHaveLength(2);
    const second = FakeWebSocket.instances[1];
    second.open();

    const frames = second.sentFrames();
    expect(frames.some((frame) => frame.type === "subscribe")).toBe(true);
    expect(
      frames.some((frame) => frame.type === "status" && frame.editing === true)
    ).toBe(true);
  });

  it("não reconecta sem assinantes ativos", async () => {
    const client = await freshClient();
    const off = client.subscribeMissionPos(params, () => {});
    const socket = FakeWebSocket.instances[0];
    socket.open();
    off();

    vi.advanceTimersByTime(30_000);
    expect(FakeWebSocket.instances).toHaveLength(1);
  });

  it("autentica por ticket como primeiro frame quando a URL direta está setada", async () => {
    vi.stubEnv("NEXT_PUBLIC_REALTIME_URL", "ws://realtime.test/realtime/ws");
    ticketFetch.mockResolvedValue({ ticket: "jwt-ticket" });
    const client = await freshClient();
    client.subscribeMissionPos(params, () => {});

    const socket = FakeWebSocket.instances[0];
    expect(socket.url).toBe("ws://realtime.test/realtime/ws");
    socket.open();
    await vi.waitFor(() => {
      expect(socket.sent.length).toBeGreaterThanOrEqual(2);
    });

    const frames = socket.sentFrames();
    expect(frames[0]).toEqual({ type: "auth", ticket: "jwt-ticket" });
    expect(frames[1]?.type).toBe("subscribe");
    expect(ticketFetch).toHaveBeenCalledWith("/realtime/ticket", {
      redirectOn401: false,
    });
  });
});
