"use client";

import { Bookmark } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
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
        variant="ghost"
        size="icon-sm"
        onClick={onClick}
        aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
        aria-pressed={selected}
        title={selected ? t("remove") : t("add")}
        disabled={!selected && !canAdd}
      >
        <Bookmark
          aria-hidden="true"
          className={selected ? "fill-primary text-primary" : "text-muted-foreground"}
        />
      </Button>
    );
  }

  return (
    <Button
      type="button"
      size={compact ? "sm" : "default"}
      onClick={onClick}
      aria-label={selected ? t("removeAria", { id: canonicalId }) : t("addAria", { id: canonicalId })}
      aria-pressed={selected}
      disabled={!selected && !canAdd}
    >
      <Bookmark
        data-icon="inline-start"
        aria-hidden="true"
        className={selected ? "fill-current" : undefined}
      />
      {t("label")}
    </Button>
  );
}

