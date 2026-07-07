import { redirect } from "next/navigation";

import { getSessionUser } from "@/services/session";

export default async function ProtectedLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getSessionUser();
  if (!user) {
    redirect("/login");
  }
  return children;
}
