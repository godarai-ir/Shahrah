"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

export type ResizableSplitPaneProps = {
  start: React.ReactNode;
  end: React.ReactNode;
  startLabel: string;
  endLabel: string;
  separatorLabel: string;
  direction?: "ltr" | "rtl";
  defaultStartSize?: number;
  minSize?: number;
  className?: string;
};

export function ResizableSplitPane({
  start,
  end,
  startLabel,
  endLabel,
  separatorLabel,
  direction = "ltr",
  defaultStartSize = 50,
  minSize = 28,
  className,
}: ResizableSplitPaneProps) {
  const containerRef = React.useRef<HTMLDivElement>(null);
  const [startSize, setStartSize] = React.useState(() => clamp(defaultStartSize, minSize, 100 - minSize));
  const draggingRef = React.useRef(false);
  const previousBodyCursorRef = React.useRef("");
  const previousBodyUserSelectRef = React.useRef("");

  const updateFromPointer = React.useCallback((clientX: number) => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect || rect.width === 0) return;
    const physical = direction === "rtl" ? rect.right - clientX : clientX - rect.left;
    setStartSize(clamp((physical / rect.width) * 100, minSize, 100 - minSize));
  }, [direction, minSize]);

  React.useEffect(() => {
    function onMove(event: PointerEvent) {
      if (draggingRef.current) updateFromPointer(event.clientX);
    }
    function onUp() {
      if (!draggingRef.current) return;
      draggingRef.current = false;
      document.body.style.cursor = previousBodyCursorRef.current;
      document.body.style.userSelect = previousBodyUserSelectRef.current;
    }
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
    window.addEventListener("pointercancel", onUp);
    window.addEventListener("blur", onUp);
    return () => {
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
      window.removeEventListener("pointercancel", onUp);
      window.removeEventListener("blur", onUp);
      if (draggingRef.current) {
        document.body.style.cursor = previousBodyCursorRef.current;
        document.body.style.userSelect = previousBodyUserSelectRef.current;
        draggingRef.current = false;
      }
    };
  }, [updateFromPointer]);

  function adjust(delta: number) {
    setStartSize((current) => clamp(current + delta, minSize, 100 - minSize));
  }

  return (
    <div
      ref={containerRef}
      dir={direction}
      className={cn("grid min-h-0 min-w-0", className)}
      style={{ gridTemplateColumns: `minmax(0, ${startSize}fr) 0.75rem minmax(0, ${100 - startSize}fr)` }}
    >
      <section aria-label={startLabel} className="min-h-0 min-w-0 overflow-hidden">{start}</section>
      <div
        role="separator"
        tabIndex={0}
        aria-label={separatorLabel}
        aria-orientation="vertical"
        aria-valuemin={minSize}
        aria-valuemax={100 - minSize}
        aria-valuenow={Math.round(startSize)}
        className="group flex cursor-col-resize touch-none items-stretch justify-center outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onPointerDown={(event) => {
          event.preventDefault();
          draggingRef.current = true;
          previousBodyCursorRef.current = document.body.style.cursor;
          previousBodyUserSelectRef.current = document.body.style.userSelect;
          document.body.style.cursor = "col-resize";
          document.body.style.userSelect = "none";
          updateFromPointer(event.clientX);
        }}
        onKeyDown={(event) => {
          const logicalStep = event.shiftKey ? 10 : 2;
          if (event.key === "Home") { event.preventDefault(); setStartSize(minSize); }
          if (event.key === "End") { event.preventDefault(); setStartSize(100 - minSize); }
          if (event.key === "ArrowLeft") { event.preventDefault(); adjust(direction === "rtl" ? logicalStep : -logicalStep); }
          if (event.key === "ArrowRight") { event.preventDefault(); adjust(direction === "rtl" ? -logicalStep : logicalStep); }
        }}
      >
        <span className="my-2 w-px bg-border transition-colors group-hover:bg-primary group-focus-visible:bg-primary" />
      </div>
      <section aria-label={endLabel} className="min-h-0 min-w-0 overflow-hidden">{end}</section>
    </div>
  );
}

function clamp(value: number, minimum: number, maximum: number) {
  return Math.min(maximum, Math.max(minimum, value));
}
