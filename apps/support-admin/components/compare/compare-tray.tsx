"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCompareModels } from "@/components/compare/compare-provider";
import { MAX_COMPARE_MODELS, serializeCompareIds } from "@/lib/compare-models";

export function CompareTray() {
  const { selectedIds, selectedCount, remove, clear, notice } = useCompareModels();

  if (selectedCount === 0) return null;

  return (
    <div className="fixed inset-x-0 bottom-0 z-30 border-t border-zinc-800 bg-zinc-950/95 backdrop-blur">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-2 px-4 py-3 md:flex-row md:items-center">
        <div className="flex items-center gap-2">
          <span className="text-sm text-zinc-300">Compare tray</span>
          <Badge>{selectedCount}/{MAX_COMPARE_MODELS}</Badge>
        </div>
        <div className="flex flex-wrap gap-1">
          {selectedIds.map((id) => (
            <button
              key={id}
              type="button"
              className="rounded border border-zinc-700 px-2 py-1 text-xs text-zinc-300 hover:border-zinc-500"
              onClick={() => remove(id)}
              aria-label={`Remove ${id} from compare`}
            >
              {id} x
            </button>
          ))}
        </div>
        <div className="ms-auto flex items-center gap-2">
          <Link
            href={`/compare?ids=${encodeURIComponent(serializeCompareIds(selectedIds))}`}
            aria-label="Open compare page"
            className="inline-flex h-8 items-center rounded-md bg-zinc-200 px-3 text-sm font-medium text-zinc-950 hover:bg-zinc-300"
          >
            Open compare
          </Link>
          <Button type="button" className="h-8 bg-zinc-800 hover:bg-zinc-700" onClick={clear} aria-label="Clear compare selection">Clear all</Button>
        </div>
      </div>
      {notice ? <p className="mx-auto max-w-7xl px-4 pb-3 text-xs text-amber-300" aria-live="polite">{notice}</p> : null}
    </div>
  );
}
