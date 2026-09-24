"use client";

import { useTranslations } from "next-intl";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { catalogEntry } from "../catalog";
import type { GuardrailType } from "../types";

type SummaryPanelProps = {
  type: GuardrailType | null;
  name: string;
  sentence: string;
};

export function SummaryPanel({ type, name, sentence }: SummaryPanelProps) {
  const t = useTranslations("guardrails");
  const Icon = type ? catalogEntry(type).icon : null;

  return (
    <Card>
      <CardHeader className="border-b border-border py-3 md:py-3">
        <CardTitle className="text-xs font-semibold text-muted-foreground">{t("wizard.summary.title")}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 pt-4 md:pt-4" aria-live="polite">
        {type && Icon ? (
          <>
            <div className="flex items-center gap-3">
              <span aria-hidden="true" className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary [&_svg]:size-4">
                <Icon />
              </span>
              <span className="flex min-w-0 flex-col">
                <span className="truncate text-sm font-semibold" dir="auto">{name.trim() || t("wizard.summary.unnamed")}</span>
                <span className="text-xs text-muted-foreground">{t(`types.${type}`)}</span>
              </span>
            </div>
            <p className="text-sm leading-7">{sentence}</p>
          </>
        ) : (
          <p className="text-sm leading-6 text-muted-foreground">{t("wizard.summary.empty")}</p>
        )}
      </CardContent>
    </Card>
  );
}
