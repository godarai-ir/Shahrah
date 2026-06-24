"use client";

import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCompareModels } from "@/components/compare/compare-provider";

type CompareToggleButtonProps = {
  canonicalId: string;
  compact?: boolean;
  iconOnly?: boolean;
};

export function CompareToggleButton({ canonicalId, compact = false, iconOnly = false }: CompareToggleButtonProps) {
  const { add, remove, isSelected, canAdd } = useCompareModels();
  const selected = isSelected(canonicalId);

  const onClick = (e?: React.MouseEvent<HTMLButtonElement>) => {
    e?.preventDefault();
    e?.stopPropagation();

    if (selected) {
      remove(canonicalId);
      return;
    }
    add(canonicalId);
  };

  if (iconOnly) {
    return (
      <Button
        type="button"
        onClick={onClick}
        aria-label={selected ? `Remove ${canonicalId} from compare` : `Add ${canonicalId} to compare`}
        title={selected ? "Remove from compare" : "Add to compare"}
        disabled={!selected && !canAdd}
        className="h-8 w-8 border border-zinc-700 bg-zinc-900/95 p-0 text-zinc-200 hover:bg-zinc-800"
      >
        {selected ? <X className="h-4 w-4 text-emerald-300" /> : <Plus className="h-4 w-4" />}
      </Button>
    );
  }

  return (
    <Button
      type="button"
      onClick={onClick}
      aria-label={selected ? `Remove ${canonicalId} from compare` : `Add ${canonicalId} to compare`}
      disabled={!selected && !canAdd}
      className={compact ? "h-8 px-2 text-xs" : ""}
    >
      {selected ? "Remove from compare" : "Add to compare"}
      {selected ? <Badge className="ml-2 border-emerald-500 text-emerald-300">Selected</Badge> : null}
    </Button>
  );
}

