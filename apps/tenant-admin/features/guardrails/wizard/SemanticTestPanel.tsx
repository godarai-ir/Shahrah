"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { CircleCheck, FlaskConical, OctagonX, Play, TriangleAlert, type LucideIcon } from "lucide-react";

import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useNumberFormatter } from "@/hooks/use-locale-format";
import { cn } from "@/lib/utils";

import { GuardrailAPIError, testSemanticGuardrail } from "../api";
import { semanticConfig, semanticConfigError } from "../semantic";
import type { GuardrailFormInput, SemanticDecision, SemanticPolicyTestResult } from "../types";

type SemanticTestPanelProps = {
  form: GuardrailFormInput;
  sample: string;
  onSampleChange: (sample: string) => void;
};

const DECISION_STYLES: Record<SemanticDecision, { icon: LucideIcon; className: string }> = {
  pass: { icon: CircleCheck, className: "bg-success-soft text-success-soft-foreground ring-success-soft-border" },
  review: { icon: TriangleAlert, className: "bg-warning-soft text-warning-soft-foreground ring-warning-soft-border" },
  block: { icon: OctagonX, className: "bg-destructive-soft text-destructive-soft-foreground ring-destructive-soft-border" },
};

function errorKey(error: unknown) {
  if (!(error instanceof GuardrailAPIError)) return "failed";
  if (error.status === 503) return "notConfigured";
  if (error.status === 429) return "rateLimited";
  if (error.status === 400) return "invalid";

  return "failed";
}

export function SemanticTestPanel({ form, sample, onSampleChange }: SemanticTestPanelProps) {
  const t = useTranslations("guardrails");
  const sampleID = useId();
  const percent = useNumberFormatter({ style: "percent", maximumFractionDigits: 0 });
  const number = useNumberFormatter({ maximumFractionDigits: 1 });
  const [result, setResult] = useState<SemanticPolicyTestResult | null>(null);
  const [error, setError] = useState("");
  const [running, setRunning] = useState(false);
  const controller = useRef<AbortController | null>(null);
  const configError = semanticConfigError(form);

  useEffect(() => () => controller.current?.abort(), []);

  async function run() {
    controller.current?.abort();
    const current = new AbortController();
    controller.current = current;
    setRunning(true);
    setError("");
    try {
      setResult(await testSemanticGuardrail({ config: semanticConfig(form), text: sample }, current.signal));
    } catch (cause) {
      if (current.signal.aborted) return;
      setResult(null);
      setError(t(`wizard.test.semantic.errors.${errorKey(cause)}`));
    } finally {
      if (controller.current === current) setRunning(false);
    }
  }

  const style = result ? DECISION_STYLES[result.decision] : null;
  const DecisionIcon = style?.icon;

  return (
    <Card>
      <CardHeader className="gap-1 border-b border-border py-3 md:py-3">
        <CardTitle className="flex items-center gap-2 text-sm font-semibold">
          <FlaskConical className="size-4" aria-hidden="true" />
          {t("wizard.test.title")}
        </CardTitle>
        <CardDescription className="text-xs leading-5">{t("wizard.test.semantic.note")}</CardDescription>
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
        <Button type="button" variant="outline" size="sm" disabled={running || !sample.trim() || configError !== null} aria-busy={running || undefined} onClick={() => void run()}>
          {running ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" aria-hidden="true" />}
          {running ? t("wizard.test.semantic.running") : t("wizard.test.semantic.run")}
        </Button>
        {configError ? <p className="text-xs leading-5 text-muted-foreground">{t("wizard.test.semantic.fixConfig")}</p> : null}

        <OperationNotice notice={error ? { kind: "error", message: error } : null} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

        {result && style && DecisionIcon ? (
          <div role="status" aria-live="polite" className="flex flex-col gap-3">
            <div className={cn("flex flex-col gap-1 rounded-lg px-3 py-2 ring-1 ring-inset", style.className)}>
              <span className="flex items-center gap-2 text-sm font-semibold">
                <DecisionIcon className="size-4 shrink-0" aria-hidden="true" />
                {t(`wizard.test.semantic.decision.${result.decision}`)}
              </span>
              <span className="text-xs leading-5">
                {t("wizard.test.semantic.meta", {
                  language: t(`wizard.test.semantic.language.${result.language}`),
                  ms: number.format(result.latency_ms),
                })}
              </span>
            </div>
            <ul className="flex flex-col gap-2">
              {result.rules.map((rule) => (
                <li key={rule.key} className="flex items-center justify-between gap-3 text-xs">
                  <span dir="ltr" className="min-w-0 truncate font-mono">{rule.key}</span>
                  <span className="flex shrink-0 items-center gap-2">
                    <span className="tabular-nums">{percent.format(rule.probability)}</span>
                    <span className="text-muted-foreground">{t(`wizard.test.semantic.decision.${rule.decision}`)}</span>
                  </span>
                </li>
              ))}
            </ul>
            {result.severity !== undefined ? (
              <p className="text-xs text-muted-foreground">{t("wizard.test.semantic.severity", { value: number.format(result.severity) })}</p>
            ) : null}
            {result.truncated ? <p className="text-xs text-muted-foreground">{t("wizard.test.semantic.truncated")}</p> : null}
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
