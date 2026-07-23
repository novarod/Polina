import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PresenceAvatars } from "@/components/missions/presence-avatars";

describe("PresenceAvatars", () => {
  it("não renderiza nada sem usuários presentes", () => {
    render(<PresenceAvatars users={[]} />);
    expect(screen.queryByTestId("presence-avatars")).not.toBeInTheDocument();
  });

  it("renderiza iniciais e nome no title de cada usuário", () => {
    render(
      <PresenceAvatars
        users={[
          { id: "u1", name: "Ana Beatriz" },
          { id: "u2", name: "caio" },
        ]}
      />
    );
    const avatars = screen.getAllByTestId("presence-avatar");
    expect(avatars).toHaveLength(2);
    expect(screen.getByTitle("Ana Beatriz")).toHaveTextContent("AB");
    expect(screen.getByTitle("caio")).toHaveTextContent("C");
  });
});
