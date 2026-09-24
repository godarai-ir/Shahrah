"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";

import { useNumberFormatter } from "@/hooks/use-locale-format";

import type { GuardrailTranslator } from "./form";
import type { SentenceFormat } from "./sentence";

/** The `guardrails` translator and locale number formatter that sentence and summary helpers expect. */
export function useGuardrailFormat(): SentenceFormat {
  const t = useTranslations("guardrails");
  const numberFormatter = useNumberFormatter();

  return useMemo(
    () => ({ t: t as unknown as GuardrailTranslator, formatNumber: (value: number) => numberFormatter.format(value) }),
    [numberFormatter, t],
  );
}
