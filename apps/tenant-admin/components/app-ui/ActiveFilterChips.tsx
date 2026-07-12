import { X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type ActiveFilterChip = { key: string; label: string; value?: string };
export type ActiveFilterChipsProps = {
  items: readonly ActiveFilterChip[];
  onRemove: (key: string) => void;
  getRemoveLabel: (item: ActiveFilterChip) => string;
  className?: string;
};

export function ActiveFilterChips({ items, onRemove, getRemoveLabel, className }: ActiveFilterChipsProps) {
  if (items.length === 0) return null;
  return (
    <div className={cn("flex flex-wrap items-center gap-2", className)}>
      {items.map((item) => (
        <Button key={item.key} type="button" variant="outline" size="sm" className="h-7 min-w-0 max-w-full rounded-full px-2 text-xs" title={getRemoveLabel(item)} aria-label={getRemoveLabel(item)} onClick={() => onRemove(item.key)}>
          <span className="min-w-0 truncate" dir="auto">{item.value ? `${item.label}=${item.value}` : item.label}</span>
          <X data-icon="inline-end" aria-hidden="true" />
        </Button>
      ))}
    </div>
  );
}
