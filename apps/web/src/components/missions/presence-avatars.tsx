"use client";

import { presenceColor, presenceInitials } from "@/lib/presence";
import type { PresenceUser } from "@/types/realtime";

interface PresenceAvatarsProps {
  users: PresenceUser[];
}

export function PresenceAvatars({ users }: PresenceAvatarsProps) {
  if (users.length === 0) {
    return null;
  }
  return (
    <div data-testid="presence-avatars" className="flex -space-x-2">
      {users.map((user) => (
        <span
          key={user.id}
          data-testid="presence-avatar"
          title={user.name}
          className="flex size-7 items-center justify-center rounded-sm border-2 border-foreground/70 font-display text-[10px] text-background shadow-[2px_2px_0_0] shadow-foreground/25"
          style={{ backgroundColor: presenceColor(user.id) }}
        >
          {presenceInitials(user.name)}
        </span>
      ))}
    </div>
  );
}
