"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { realtimeClient } from "@/services/realtime";
import type {
  CursorMove,
  MissionTopicParams,
  PresenceUser,
} from "@/types/realtime";

interface UseMissionPresenceOptions extends MissionTopicParams {
  editing: boolean;
}

export interface MissionPresence {
  self: PresenceUser | null;
  users: PresenceUser[];
  peers: PresenceUser[];
  subscribeCursor: (listener: (move: CursorMove) => void) => () => void;
  sendPos: (x: number, y: number) => void;
}

export function useMissionPresence({
  orgId,
  workspaceId,
  missionId,
  editing,
}: UseMissionPresenceOptions): MissionPresence {
  const [self, setSelf] = useState<PresenceUser | null>(null);
  const [users, setUsers] = useState<PresenceUser[]>([]);
  const selfRef = useRef<PresenceUser | null>(null);
  const cursorListeners = useRef(new Set<(move: CursorMove) => void>());

  useEffect(() => {
    const offSelf = realtimeClient.onSelf((user) => {
      selfRef.current = user;
      setSelf(user);
    });
    const off = realtimeClient.subscribeMissionPos(
      { orgId, workspaceId, missionId },
      (frame) => {
        if (frame.type === "snapshot") {
          setUsers(frame.users ?? []);
          return;
        }
        if (frame.type === "join" && frame.user) {
          const joined = frame.user;
          setUsers((prev) =>
            prev.some((user) => user.id === joined.id)
              ? prev
              : [...prev, joined]
          );
          return;
        }
        if (frame.type === "leave" && frame.user) {
          const left = frame.user;
          setUsers((prev) => prev.filter((user) => user.id !== left.id));
          return;
        }
        if (
          frame.type === "pos" &&
          frame.user_id &&
          frame.user_id !== selfRef.current?.id
        ) {
          const move: CursorMove = {
            userId: frame.user_id,
            x: frame.x ?? 0,
            y: frame.y ?? 0,
          };
          cursorListeners.current.forEach((listener) => listener(move));
        }
      }
    );
    return () => {
      off();
      offSelf();
      setUsers([]);
    };
  }, [orgId, workspaceId, missionId]);

  useEffect(() => {
    if (!editing) {
      return;
    }
    realtimeClient.setEditing({ orgId, workspaceId, missionId }, true);
    return () => {
      realtimeClient.setEditing({ orgId, workspaceId, missionId }, false);
    };
  }, [orgId, workspaceId, missionId, editing]);

  const subscribeCursor = useCallback(
    (listener: (move: CursorMove) => void) => {
      cursorListeners.current.add(listener);
      return () => {
        cursorListeners.current.delete(listener);
      };
    },
    []
  );

  const sendPos = useCallback(
    (x: number, y: number) => {
      realtimeClient.sendPos(missionId, x, y);
    },
    [missionId]
  );

  const peers = useMemo(
    () => users.filter((user) => user.id !== self?.id),
    [users, self]
  );

  return { self, users, peers, subscribeCursor, sendPos };
}
