import { cache } from "react";

import { serverFetch } from "@/services/server-api";
import type { SessionUser } from "@/types/auth";

export const getSessionUser = cache((): Promise<SessionUser | null> => {
  return serverFetch<SessionUser>("/auth/me");
});
