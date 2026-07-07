export type Role = "VIEWER" | "DESIGNER" | "ADMIN";

const roleOrder: Record<Role, number> = {
  VIEWER: 0,
  DESIGNER: 1,
  ADMIN: 2,
};

export function roleAtLeast(role: Role, min: Role): boolean {
  return roleOrder[role] >= roleOrder[min];
}
