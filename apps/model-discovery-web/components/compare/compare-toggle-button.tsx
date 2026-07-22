"use client";

import { Plus, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCompareModels } from "@/components/compare/compare-provider";

type CompareToggleButtonProps = {
  canonicalId: string;
  compact?: boolean;
  iconOnly?: boolean;
};

export function CompareToggleButton({ canonicalId, compact = false, iconOnly = false }: CompareToggleButtonProps) {
  const t = useTranslations("Compare");
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
        aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
        title={selected ? t("remove") : t("add")}
        disabled={!selected && !canAdd}
        className="h-8 w-8 border border-border bg-card p-0 text-card-foreground hover:bg-accent"
      >
        {selected ? <X className="h-4 w-4 text-success-soft-foreground" /> : <Plus className="h-4 w-4" />}
      </Button>
    );
  }

  return (
    <Button
      type="button"
      onClick={onClick}
      aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
      disabled={!selected && !canAdd}
      className={compact ? "h-8 px-2 text-xs" : ""}
    >
      {selected ? t("remove") : t("add")}
      {selected ? <Badge className="ms-2 border-success-soft bg-success-soft text-success-soft-foreground">{t("selected")}</Badge> : null}
    </Button>
  );
}

