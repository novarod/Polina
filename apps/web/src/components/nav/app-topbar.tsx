import Link from "next/link";

import { UserMenu } from "@/components/nav/user-menu";

export function AppTopbar({ userName }: { userName: string }) {
  return (
    <header className="flex items-center justify-between border-b bg-card px-4 py-2">
      <Link href="/orgs" className="font-display text-sm text-primary">
        Polina
      </Link>
      <UserMenu name={userName} />
    </header>
  );
}
