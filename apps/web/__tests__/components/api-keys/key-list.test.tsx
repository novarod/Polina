import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { KeyList } from "@/components/api-keys/key-list";
import { revokeApiKey } from "@/services/api-keys";
import type { ApiKey } from "@/types/api-key";

vi.mock("@/services/api-keys");

const refresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

const keys: ApiKey[] = [
  {
    id: "k1",
    name: "Plugin UE5",
    last_used_at: "2026-07-08T10:00:00Z",
    created_at: "2026-07-01T10:00:00Z",
    revoked_at: null,
  },
  {
    id: "k2",
    name: "Chave antiga",
    last_used_at: null,
    created_at: "2026-06-01T10:00:00Z",
    revoked_at: "2026-07-01T10:00:00Z",
  },
];

describe("KeyList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("dims revoked keys and hides their revoke action", () => {
    render(<KeyList orgId="o1" keys={keys} />);

    const rows = screen.getAllByTestId("key-row");
    expect(rows[1].className).toContain("opacity-50");
    expect(screen.getByTestId("key-revoked")).toBeVisible();
    expect(
      screen.queryByLabelText("Revogar Chave antiga")
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText("Revogar Plugin UE5")).toBeVisible();
  });

  it("revokes an active key after confirmation", async () => {
    vi.mocked(revokeApiKey).mockResolvedValue(undefined);
    const user = userEvent.setup();
    render(<KeyList orgId="o1" keys={keys} />);

    await user.click(screen.getByLabelText("Revogar Plugin UE5"));
    expect(
      screen.getByText("Revogar a chave “Plugin UE5”?")
    ).toBeVisible();
    await user.click(screen.getByRole("button", { name: "Revogar" }));

    await waitFor(() => {
      expect(revokeApiKey).toHaveBeenCalledWith("o1", "k1");
      expect(refresh).toHaveBeenCalled();
    });
  });
});
