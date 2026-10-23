"use client";

import { useCallback } from "react";
import { useTranslations } from "next-intl";

import { useNumberFormatter } from "@/hooks/use-locale-format";

import type { PolicyRule } from "../policy-rules";
import { ruleSentence, type SentenceFormat } from "../rule-sentence";

/** The rule sentence formatter bound to the current locale. */
export function useRuleSentence(): (rule: PolicyRule) => string {
  const t = useTranslations("gatewayPolicies");
  const numberFormatter = useNumberFormatter({ maximumFractionDigits: 2 });
  return useCallback(
    (rule: PolicyRule) => ruleSentence(rule, { t: t as SentenceFormat["t"], formatNumber: (value) => numberFormatter.format(value) }),
    [numberFormatter, t],
  );
}

/** Live preview of the sentence the list will show for the rule being edited. */
export function RuleSentencePreview({ rule }: { rule: PolicyRule }) {
  const t = useTranslations("gatewayPolicies");
  const sentence = useRuleSentence();
  return (
    <div className="rounded-lg border border-border bg-muted/40 px-4 py-3">
      <p className="text-xs text-muted-foreground">{t("editor.preview")}</p>
      <p className="mt-1 text-sm leading-6 font-medium">{sentence(rule)}</p>
    </div>
  );
}
