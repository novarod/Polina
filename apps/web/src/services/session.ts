import { cache } from "react";
import { cookies } from "next/headers";

import { env } from "@/lib/env";
import type { SessionUser } from "@/types/auth";

export const getSessionUser = cache(async (): Promise<SessionUser | null> => {
  const cookieStore = await cookies();
  const session = cookieStore.get("session");
  if (!session) {
    return null;
  }

  let response: Response;
  try {
    response = await fetch(`${env.apiUrl}/auth/me`, {
      headers: { cookie: `session=${session.value}` },
      cache: "no-store",
    });
  } catch {
    return null;
  }

  if (!response.ok) {
    return null;
  }
  return (await response.json()) as SessionUser;
});
