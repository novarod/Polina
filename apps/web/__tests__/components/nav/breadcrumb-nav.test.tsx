import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";

describe("BreadcrumbNav", () => {
  it("renders links for intermediate items and plain text for the current page", () => {
    render(
      <BreadcrumbNav
        items={[
          { label: "Organizações", href: "/orgs" },
          { label: "Acme", href: "/orgs/o1" },
          { label: "DLC de Natal" },
        ]}
      />
    );

    const orgLink = screen.getByRole("link", { name: "Organizações" });
    expect(orgLink).toHaveAttribute("href", "/orgs");
    expect(screen.getByRole("link", { name: "Acme" })).toHaveAttribute(
      "href",
      "/orgs/o1"
    );
    expect(screen.getByText("DLC de Natal")).not.toHaveAttribute("href");
  });
});
