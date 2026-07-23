import { apiFetch } from "@/services/api";
import type {
  MissionTopicParams,
  PresenceUser,
  RealtimeServerFrame,
} from "@/types/realtime";

const RECONNECT_BASE_MS = 500;
const RECONNECT_MAX_MS = 10_000;

type FrameListener = (frame: RealtimeServerFrame) => void;

interface TopicSubscription {
  refCount: number;
  subscribeFrame: Record<string, unknown>;
  unsubscribeFrame: Record<string, unknown>;
  listeners: Set<FrameListener>;
}

function posTopic(missionId: string): string {
  return `pos:mission:${missionId}`;
}

function statusTopic(orgId: string): string {
  return `status:org:${orgId}`;
}

class RealtimeClient {
  private socket: WebSocket | null = null;
  private ready = false;
  private retry = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private topics = new Map<string, TopicSubscription>();
  private editing = new Map<string, Record<string, unknown>>();
  private self: PresenceUser | null = null;
  private selfListeners = new Set<(user: PresenceUser) => void>();

  subscribeMissionPos(
    params: MissionTopicParams,
    listener: FrameListener
  ): () => void {
    return this.subscribeTopic(
      posTopic(params.missionId),
      {
        type: "subscribe",
        plane: "pos",
        org_id: params.orgId,
        workspace_id: params.workspaceId,
        mission_id: params.missionId,
      },
      { type: "unsubscribe", plane: "pos", mission_id: params.missionId },
      listener
    );
  }

  subscribeOrgStatus(orgId: string, listener: FrameListener): () => void {
    return this.subscribeTopic(
      statusTopic(orgId),
      { type: "subscribe", plane: "status", org_id: orgId },
      { type: "unsubscribe", plane: "status", org_id: orgId },
      listener
    );
  }

  onSelf(listener: (user: PresenceUser) => void): () => void {
    this.selfListeners.add(listener);
    if (this.self) {
      listener(this.self);
    }
    return () => {
      this.selfListeners.delete(listener);
    };
  }

  sendPos(missionId: string, x: number, y: number): void {
    this.sendFrame({ type: "pos", mission_id: missionId, x, y });
  }

  setEditing(params: MissionTopicParams, editing: boolean): void {
    const frame = {
      type: "status",
      org_id: params.orgId,
      workspace_id: params.workspaceId,
      mission_id: params.missionId,
      editing,
    };
    if (editing) {
      this.editing.set(params.missionId, frame);
      this.sendFrame(frame);
      this.ensureSocket();
      return;
    }
    this.editing.delete(params.missionId);
    this.sendFrame(frame);
    this.maybeDisconnect();
  }

  private subscribeTopic(
    topic: string,
    subscribeFrame: Record<string, unknown>,
    unsubscribeFrame: Record<string, unknown>,
    listener: FrameListener
  ): () => void {
    let subscription = this.topics.get(topic);
    if (!subscription) {
      subscription = {
        refCount: 0,
        subscribeFrame,
        unsubscribeFrame,
        listeners: new Set(),
      };
      this.topics.set(topic, subscription);
      this.sendFrame(subscribeFrame);
      this.ensureSocket();
    }
    subscription.refCount++;
    subscription.listeners.add(listener);
    let active = true;
    return () => {
      if (!active) {
        return;
      }
      active = false;
      subscription.listeners.delete(listener);
      subscription.refCount--;
      if (subscription.refCount > 0) {
        return;
      }
      this.topics.delete(topic);
      this.sendFrame(subscription.unsubscribeFrame);
      this.maybeDisconnect();
    };
  }

  private ensureSocket(): void {
    if (this.socket || typeof window === "undefined") {
      return;
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    const directUrl = process.env.NEXT_PUBLIC_REALTIME_URL;
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url =
      directUrl ?? `${wsProtocol}//${window.location.host}/api/realtime/ws`;
    const socket = new WebSocket(url);
    this.socket = socket;
    socket.onopen = () => {
      void this.handleOpen(socket, Boolean(directUrl));
    };
    socket.onmessage = (event) => {
      this.handleMessage(String(event.data));
    };
    socket.onclose = () => {
      this.handleClose(socket);
    };
    socket.onerror = () => {};
  }

  private async handleOpen(
    socket: WebSocket,
    needsTicket: boolean
  ): Promise<void> {
    if (needsTicket) {
      try {
        const { ticket } = await apiFetch<{ ticket: string }>(
          "/realtime/ticket",
          { redirectOn401: false }
        );
        if (this.socket !== socket || socket.readyState !== WebSocket.OPEN) {
          return;
        }
        socket.send(JSON.stringify({ type: "auth", ticket }));
      } catch {
        socket.close();
        return;
      }
    }
    this.ready = true;
    this.flush();
  }

  private handleMessage(data: string): void {
    let frame: RealtimeServerFrame;
    try {
      frame = JSON.parse(data) as RealtimeServerFrame;
    } catch {
      return;
    }
    if (frame.type === "hello" && frame.user) {
      this.retry = 0;
      this.self = frame.user;
      const user = frame.user;
      this.selfListeners.forEach((listener) => listener(user));
      return;
    }
    if (!frame.topic) {
      return;
    }
    const subscription = this.topics.get(frame.topic);
    subscription?.listeners.forEach((listener) => listener(frame));
  }

  private handleClose(socket: WebSocket): void {
    if (this.socket !== socket) {
      return;
    }
    this.socket = null;
    this.ready = false;
    if (this.topics.size === 0 && this.editing.size === 0) {
      return;
    }
    const delay = Math.min(
      RECONNECT_BASE_MS * 2 ** this.retry,
      RECONNECT_MAX_MS
    );
    this.retry++;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.ensureSocket();
    }, delay);
  }

  private flush(): void {
    this.topics.forEach((subscription) => {
      this.sendFrame(subscription.subscribeFrame);
    });
    this.editing.forEach((frame) => {
      this.sendFrame(frame);
    });
  }

  private sendFrame(frame: Record<string, unknown>): void {
    if (
      !this.ready ||
      !this.socket ||
      this.socket.readyState !== WebSocket.OPEN
    ) {
      return;
    }
    this.socket.send(JSON.stringify(frame));
  }

  private maybeDisconnect(): void {
    if (this.topics.size > 0 || this.editing.size > 0) {
      return;
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.retry = 0;
    if (this.socket) {
      const socket = this.socket;
      this.socket = null;
      this.ready = false;
      socket.close(1000);
    }
  }
}

export const realtimeClient = new RealtimeClient();
