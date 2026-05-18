"use client";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCompareModels } from "@/components/compare/compare-provider";

type CompareToggleButtonProps = {
  canonicalId: string;
  compact?: boolean;
};

export function CompareToggleButton({ canonicalId, compact = false }: CompareToggleButtonProps) {
  const { add, remove, isSelected, canAdd } = useCompareModels();
  const selected = isSelected(canonicalId);

  const onClick = () => {
    if (selected) {
      remove(canonicalId);
      return;
    }
    add(canonicalId);
  };

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