import { act, render, screen, waitFor } from "@testing-library/react";
import { ReactFlow } from "@xyflow/react";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { PresenceCursors } from "@/components/canvas/presence-cursors";
import type { CursorMove } from "@/types/realtime";

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

beforeAll(() => {
  globalThis.ResizeObserver = globalThis.ResizeObserver ?? ResizeObserverStub;
});

let cursorListener: ((move: CursorMove) => void) | null = null;
const subscribeCursor = vi.fn((listener: (move: CursorMove) => void) => {
  cursorListener = listener;
  return () => {
    cursorListener = null;
  };
});
const sendPos = vi.fn();

const peers = [
  { id: "peer-1", name: "Outra Pessoa" },
  { id: "peer-2", name: "Mais Uma" },
];

function renderOverlay(overlayPeers = peers) {
  return render(
    <div style={{ width: 800, height: 600 }}>
      <ReactFlow nodes={[]} edges={[]}>
        <PresenceCursors
          peers={overlayPeers}
          subscribeCursor={subscribeCursor}
          sendPos={sendPos}
        />
      </ReactFlow>
    </div>
  );
}

describe("PresenceCursors", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cursorListener = null;
  });

  it("renderiza um cursor por peer com nome e sem pointer events", () => {
    renderOverlay();
    const cursors = screen.getAllByTestId("peer-cursor");
    expect(cursors).toHaveLength(2);
    expect(screen.getByText("Outra Pessoa")).toBeInTheDocument();
    expect(cursors[0].className).toContain("pointer-events-none");
  });

  it("começa fora da viewport e move para a posição recebida", async () => {
    renderOverlay();
    const cursor = screen
      .getAllByTestId("peer-cursor")
      .find((el) => el.dataset.userId === "peer-1");
    expect(cursor?.style.transform).toContain("-9999px");

    expect(cursorListener).not.toBeNull();
    act(() => {
      cursorListener?.({ userId: "peer-1", x: 120, y: 80 });
    });

    await waitFor(() => {
      expect(cursor?.style.transform).toBe("translate(120px, 80px)");
    });
  });

  it("não renderiza nada sem peers", () => {
    renderOverlay([]);
    expect(screen.queryByTestId("peer-cursor")).not.toBeInTheDocument();
  });

  it("cancela a assinatura de cursores ao desmontar", () => {
    const { unmount } = renderOverlay();
    expect(cursorListener).not.toBeNull();
    unmount();
    expect(cursorListener).toBeNull();
  });
});
