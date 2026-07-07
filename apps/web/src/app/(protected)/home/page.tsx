import type { Metadata } from "next";

import { LogoutButton } from "@/components/auth/logout-button";
import { getSessionUser } from "@/services/session";

export const metadata: Metadata = {
  title: "Home — Polina",
};

export default async function HomePage() {
  const user = await getSessionUser();

  return (
    <main className="flex min-h-dvh flex-col items-center justify-center gap-6 p-4">
      <h1 className="font-display text-2xl text-primary">Polina</h1>
      <p data-testid="session-user" className="text-lg">
        Olá, <span className="font-medium">{user?.name}</span>
      </p>
      <LogoutButton />
    </main>
  );
}
