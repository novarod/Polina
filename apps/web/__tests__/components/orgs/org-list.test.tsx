import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { OrgList } from "@/components/orgs/org-list";
import { WorkspaceList } from "@/components/workspaces/workspace-list";
import type { OrganizationListItem } from "@/types/organization";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: vi.fn() }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

function org(role: OrganizationListItem["role"]): OrganizationListItem {
  return { id: "o1", name: "Acme", slug: "acme", role };
}

describe("role gating", () => {
  it("hides org rename/delete from a DESIGNER", () => {
    render(<OrgList orgs={[org("DESIGNER")]} />);

    expect(screen.queryByLabelText("Renomear Acme")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Deletar Acme")).not.toBeInTheDocument();
    expect(screen.getByTestId("create-org")).toBeVisible();
  });

  it("shows org rename/delete to an ADMIN", () => {
    render(<OrgList orgs={[org("ADMIN")]} />);

    expect(screen.getByLabelText("Renomear Acme")).toBeVisible();
    expect(screen.getByLabelText("Deletar Acme")).toBeVisible();
  });

  it("hides workspace creation and actions from a VIEWER", () => {
    render(
      <WorkspaceList
        orgId="o1"
        role="VIEWER"
        workspaces={[
          {
            id: "w1",
            organization_id: "o1",
            name: "Sistemas",
            description: "",
            created_at: "2026-07-07T00:00:00Z",
          },
        ]}
      />
    );

    expect(screen.queryByTestId("create-workspace")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Editar Sistemas")).not.toBeInTheDocument();
    expect(
      screen.queryByLabelText("Deletar Sistemas")
    ).not.toBeInTheDocument();
    expect(screen.getByText("Sistemas")).toBeVisible();
  });

  it("shows empty state without creation CTA to a VIEWER", () => {
    render(<WorkspaceList orgId="o1" role="VIEWER" workspaces={[]} />);

    expect(screen.getByText("Nenhum workspace ainda.")).toBeVisible();
    expect(screen.queryByTestId("create-workspace")).not.toBeInTheDocument();
  });
});
