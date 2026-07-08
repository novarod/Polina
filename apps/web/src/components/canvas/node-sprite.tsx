const sprites: Record<string, React.ReactNode> = {
  START: (
    <>
      <rect x="4" y="3" width="2" height="10" />
      <rect x="6" y="4" width="2" height="8" />
      <rect x="8" y="5" width="2" height="6" />
      <rect x="10" y="6" width="2" height="4" />
      <rect x="12" y="7" width="2" height="2" />
    </>
  ),
  END: (
    <>
      <rect x="3" y="2" width="2" height="12" />
      <rect x="5" y="2" width="2" height="2" />
      <rect x="9" y="2" width="2" height="2" />
      <rect x="7" y="4" width="2" height="2" />
      <rect x="11" y="4" width="2" height="2" />
      <rect x="5" y="6" width="2" height="2" />
      <rect x="9" y="6" width="2" height="2" />
    </>
  ),
  DIALOGUE: (
    <>
      <rect x="2" y="3" width="12" height="8" />
      <rect x="5" y="11" width="2" height="2" />
      <rect x="4" y="13" width="1" height="1" />
    </>
  ),
  KILL: (
    <>
      <rect x="11" y="2" width="2" height="2" />
      <rect x="9" y="4" width="2" height="2" />
      <rect x="7" y="6" width="2" height="2" />
      <rect x="5" y="8" width="2" height="2" />
      <rect x="3" y="9" width="6" height="2" />
      <rect x="3" y="11" width="2" height="3" />
    </>
  ),
  COLLECT: (
    <>
      <rect x="2" y="3" width="12" height="4" />
      <rect x="2" y="8" width="12" height="5" />
      <rect x="7" y="7" width="2" height="3" />
    </>
  ),
  OBJECTIVE: (
    <>
      <rect x="7" y="2" width="2" height="8" />
      <rect x="7" y="12" width="2" height="2" />
    </>
  ),
};

const fallbackSprite = (
  <>
    <rect x="5" y="3" width="6" height="2" />
    <rect x="9" y="5" width="2" height="2" />
    <rect x="7" y="7" width="2" height="2" />
    <rect x="7" y="11" width="2" height="2" />
  </>
);

export function NodeSprite({ type }: { type: string }) {
  return (
    <svg
      viewBox="0 0 16 16"
      className="size-6 shrink-0 fill-current"
      shapeRendering="crispEdges"
      aria-hidden="true"
      data-testid="node-sprite"
      data-sprite={type in sprites ? type : "UNKNOWN"}
    >
      {sprites[type] ?? fallbackSprite}
    </svg>
  );
}
