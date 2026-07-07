import { cookies } from "next/headers";

import { env } from "@/lib/env";

export async function serverFetch<T>(path: string): Promise<T | null> {
  const cookieStore = await cookies();
  const session = cookieStore.get("session");
  if (!session) {
    return null;
  }

  let response: Response;
  try {
    response = await fetch(`${env.apiUrl}${path}`, {
      headers: { cookie: `session=${session.value}` },
      cache: "no-store",
    });
  } catch {
    return null;
  }

  if (!response.ok) {
    return null;
  }
  return (await response.json()) as T;
}
