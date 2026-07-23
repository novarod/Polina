"use client";

import { useEffect, useState } from "react";

import { realtimeClient } from "@/services/realtime";

export function useOrgStatus(orgId: string): Record<string, number> {
  const [counts, setCounts] = useState<Record<string, number>>({});

  useEffect(() => {
    const off = realtimeClient.subscribeOrgStatus(orgId, (frame) => {
      if (frame.type === "snapshot") {
        const next: Record<string, number> = {};
        for (const entry of frame.missions ?? []) {
          if (entry.count > 0) {
            next[entry.mission_id] = entry.count;
          }
        }
        setCounts(next);
        return;
      }
      if (frame.type === "status" && frame.mission_id) {
        const missionId = frame.mission_id;
        const count = frame.count ?? 0;
        setCounts((prev) => {
          const next = { ...prev };
          if (count > 0) {
            next[missionId] = count;
          } else {
            delete next[missionId];
          }
          return next;
        });
      }
    });
    return () => {
      off();
      setCounts({});
    };
  }, [orgId]);

  return counts;
}
