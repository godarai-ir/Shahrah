"use client";

import { useId, useMemo } from "react";
import { useTranslations } from "next-intl";
import { CircleCheck, Eraser, Flag, FlaskConical, Info, OctagonX, type LucideIcon } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

import { type EvaluationDetail, type EvaluationStatus, evaluateSample } from "../evaluate";
import { catalogEntry } from "../catalog";
import type { SentenceFormat } from "../sentence";
import type { GuardrailFormInput } from "../types";

type TestPanelProps = {
  form: GuardrailFormInput;
  sample: string;
  format: SentenceFormat;
  onSampleChange: (sample: string) => void;
};

const STATUS_STYLES: Record<EvaluationStatus, { icon: LucideIcon; className: string }> = {
  pass: { icon: CircleCheck, className: "bg-success-soft text-success-soft-foreground ring-success-soft-border" },
  mutate: { icon: Eraser, className: "bg-primary-soft text-primary-soft-foreground ring-primary-subtle-border" },
  flag: { icon: Flag, className: "bg-warning-soft text-warning-soft-foreground ring-warning-soft-border" },
  block: { icon: OctagonX, className: "bg-destructive-soft text-destructive-soft-foreground ring-destructive-soft-border" },
  blockNoLog: { icon: OctagonX, className: "bg-destructive-soft text-destructive-soft-foreground ring-destructive-soft-border" },
  na: { icon: Info, className: "bg-muted text-muted-foreground ring-border" },
};

function detailText(detail: EvaluationDetail, form: GuardrailFormInput, { t, formatNumber }: SentenceFormat) {
  const values = {
    n: detail.count ?? 0,
    count: formatNumber(detail.count ?? 0),
    total: formatNumber(detail.total ?? 0),
    items: (detail.items ?? []).join(t("sentence.listSeparator")),
  };
  if (detail.key === "count") return t(`wizard.test.details.count.${catalogEntry(form.type).unit ?? "word"}`, values);

  return t(`wizard.test.details.${detail.key}`, values);
}

export function TestPanel({ form, sample, format, onSampleChange }: TestPanelProps) {
  const t = useTranslations("guardrails");
  const sampleID = useId();
  const result = useMemo(() => evaluateSample(form, sample), [form, sample]);
  const style = STATUS_STYLES[result.status];
  const StatusIcon = style.icon;

  return (
    <Card>
      <CardHeader className="gap-1 border-b border-border py-3 md:py-3">
        <CardTitle className="flex items-center gap-2 text-sm font-semibold">
          <FlaskConical className="size-4" aria-hidden="true" />
          {t("wizard.test.title")}
        </CardTitle>
        <CardDescription className="text-xs leading-5">{t("wizard.test.note")}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 pt-4 md:pt-4">
        <label htmlFor={sampleID} className="sr-only">{t("wizard.test.sampleLabel")}</label>
        <Textarea
          id={sampleID}
          dir="auto"
          rows={4}
          value={sample}
          placeholder={t("wizard.test.placeholder")}
          onChange={(event) => onSampleChange(event.target.value)}
          className="min-h-24 text-sm leading-6"
        />
        <div role="status" aria-live="polite" className={cn("flex flex-col gap-1 rounded-lg px-3 py-2 ring-1 ring-inset", style.className)}>
          <span className="flex items-center gap-2 text-sm font-semibold">
            <StatusIcon className="size-4 shrink-0" aria-hidden="true" />
            {t(`wizard.test.result.${result.status}`)}
          </span>
          <span className="break-words text-xs leading-5">{detailText(result.detail, form, format)}</span>
        </div>
        {result.status === "mutate" && result.output !== undefined ? (
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t("wizard.test.output")}</span>
            <div dir="auto" className="whitespace-pre-wrap break-words rounded-lg border border-code-border bg-code px-3 py-2 text-sm leading-6 text-code-foreground">
              {result.output}
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
