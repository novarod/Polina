import Link from "next/link";

import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center gap-4 p-4">
      <h1 className="font-display text-lg text-primary">404</h1>
      <p className="text-muted-foreground">Página não encontrada.</p>
      <Button asChild variant="outline">
        <Link href="/orgs">Voltar para as organizações</Link>
      </Button>
    </main>
  );
}
