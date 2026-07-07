"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronDownIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { logout } from "@/services/auth";

export function UserMenu({ name }: { name: string }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function onLogout() {
    setPending(true);
    try {
      await logout();
      router.push("/login");
      router.refresh();
    } finally {
      setPending(false);
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" data-testid="user-menu">
          {name}
          <ChevronDownIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          variant="destructive"
          disabled={pending}
          onSelect={onLogout}
        >
          {pending ? "Saindo..." : "Sair"}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
