"use client";

import { useEffect, useRef } from "react";
import {
  ViewportPortal,
  useReactFlow,
  useStore,
  useViewport,
} from "@xyflow/react";

import { presenceColor } from "@/lib/presence";
import type { CanvasPresence } from "@/types/realtime";

const SEND_INTERVAL_MS = 100;
const LERP_FACTOR = 0.35;
const SETTLE_DISTANCE = 0.5;

export function PresenceCursors({
  peers,
  subscribeCursor,
  sendPos,
}: CanvasPresence) {
  const { screenToFlowPosition } = useReactFlow();
  const domNode = useStore((state) => state.domNode);
  const { zoom } = useViewport();
  const cursorRefs = useRef(new Map<string, HTMLDivElement>());
  const targets = useRef(new Map<string, { x: number; y: number }>());

  useEffect(() => {
    if (!domNode) {
      return;
    }
    let lastSent = 0;
    let raf = 0;
    let pending: { x: number; y: number } | null = null;
    const handler = (event: PointerEvent) => {
      pending = { x: event.clientX, y: event.clientY };
      if (raf) {
        return;
      }
      raf = requestAnimationFrame(() => {
        raf = 0;
        if (!pending) {
          return;
        }
        const now = performance.now();
        if (now - lastSent < SEND_INTERVAL_MS) {
          return;
        }
        lastSent = now;
        const point = screenToFlowPosition(pending);
        sendPos(point.x, point.y);
      });
    };
    domNode.addEventListener("pointermove", handler);
    return () => {
      domNode.removeEventListener("pointermove", handler);
      if (raf) {
        cancelAnimationFrame(raf);
      }
    };
  }, [domNode, screenToFlowPosition, sendPos]);

  useEffect(() => {
    let raf = 0;
    const current = new Map<string, { x: number; y: number }>();
    const tick = () => {
      raf = 0;
      let settling = false;
      targets.current.forEach((target, userId) => {
        const element = cursorRefs.current.get(userId);
        if (!element) {
          return;
        }
        const position = current.get(userId) ?? target;
        const x = position.x + (target.x - position.x) * LERP_FACTOR;
        const y = position.y + (target.y - position.y) * LERP_FACTOR;
        current.set(userId, { x, y });
        element.style.transform = `translate(${x}px, ${y}px)`;
        if (
          Math.abs(target.x - x) > SETTLE_DISTANCE ||
          Math.abs(target.y - y) > SETTLE_DISTANCE
        ) {
          settling = true;
        }
      });
      if (settling) {
        raf = requestAnimationFrame(tick);
      }
    };
    const unsubscribe = subscribeCursor(({ userId, x, y }) => {
      targets.current.set(userId, { x, y });
      if (!raf) {
        raf = requestAnimationFrame(tick);
      }
    });
    return () => {
      unsubscribe();
      if (raf) {
        cancelAnimationFrame(raf);
      }
    };
  }, [subscribeCursor]);

  return (
    <ViewportPortal>
      {peers.map((peer) => {
        const color = presenceColor(peer.id);
        return (
          <div
            key={peer.id}
            data-testid="peer-cursor"
            data-user-id={peer.id}
            ref={(element) => {
              if (element) {
                cursorRefs.current.set(peer.id, element);
              } else {
                cursorRefs.current.delete(peer.id);
                targets.current.delete(peer.id);
              }
            }}
            className="pointer-events-none absolute top-0 left-0 z-10 will-change-transform"
            style={{ transform: "translate(-9999px, -9999px)" }}
          >
            <div
              className="flex items-start"
              style={{ transform: `scale(${1 / zoom})`, transformOrigin: "top left" }}
            >
              <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden>
                <path
                  d="M2 1l5 13 2-5.5L14.5 6z"
                  fill={color}
                  stroke="var(--foreground)"
                  strokeWidth="1.5"
                  strokeLinejoin="round"
                />
              </svg>
              <span
                className="mt-3 -ml-1 rounded-sm border border-foreground/40 px-1 font-display text-[8px] leading-3 whitespace-nowrap text-background"
                style={{ backgroundColor: color }}
              >
                {peer.name}
              </span>
            </div>
          </div>
        );
      })}
    </ViewportPortal>
  );
}
