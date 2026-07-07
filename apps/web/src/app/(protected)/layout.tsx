import { redirect } from "next/navigation";

import { AppTopbar } from "@/components/nav/app-topbar";
import { getSessionUser } from "@/services/session";

export default async function ProtectedLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getSessionUser();
  if (!user) {
    redirect("/login");
  }
  return (
    <div className="flex min-h-dvh flex-col">
      <AppTopbar userName={user.name} />
      {children}
    </div>
  );
}
