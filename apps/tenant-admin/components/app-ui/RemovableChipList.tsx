import * as React from "react";
import { X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type RemovableChipItem = { value: string; label?: React.ReactNode; removeLabel: string; textDir?: "rtl" | "ltr" | "auto" };
export type RemovableChipListProps = { items: readonly RemovableChipItem[]; onRemove: (value: string) => void; disabled?: boolean; variant?: React.ComponentProps<typeof Badge>["variant"]; className?: string };

export function RemovableChipList({ items, onRemove, disabled, variant = "secondary", className }: RemovableChipListProps) {
  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {items.map((item) => (
        <Badge key={item.value} variant={variant} className="gap-1 pe-1 font-normal">
          <span className="max-w-64 truncate" dir={item.textDir}>{item.label ?? item.value}</span>
          <button type="button" className="rounded-sm text-muted-foreground hover:text-foreground" onClick={() => onRemove(item.value)} aria-label={item.removeLabel} disabled={disabled}><X aria-hidden="true" /></button>
        </Badge>
      ))}
    </div>
  );
}
