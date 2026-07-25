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
        variant="outline"
        size="icon-sm"
        onClick={onClick}
        aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
        title={selected ? t("remove") : t("add")}
        disabled={!selected && !canAdd}
      >
        {selected ? (
          <X data-icon="inline-start" className="text-success-soft-foreground" />
        ) : (
          <Plus data-icon="inline-start" />
        )}
      </Button>
    );
  }

  return (
    <Button
      type="button"
      size={compact ? "sm" : "default"}
      onClick={onClick}
      aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
      disabled={!selected && !canAdd}
    >
      {selected ? t("remove") : t("add")}
      {selected ? <Badge className="ms-2 border-success-soft bg-success-soft text-success-soft-foreground">{t("selected")}</Badge> : null}
    </Button>
  );
}

