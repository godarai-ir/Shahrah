"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import { TriangleAlert } from "lucide-react";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { useNumberFormatter } from "@/hooks/use-locale-format";
import { cn } from "@/lib/utils";

import { MAX_SEMANTIC_SEVERITY, SEMANTIC_PACKS, validThresholds } from "../semantic";
import type { GuardrailEnforcement, GuardrailFormInput, SemanticThresholdInput } from "../types";

type SemanticThresholdPatch = Partial<
  Pick<GuardrailFormInput, "semanticPackRules" | "semanticRules" | "semanticFinglish" | "semanticFinglishThresholds" | "semanticSeverity" | "semanticSeverityBlockAt">
>;

type SemanticThresholdsProps = {
  form: GuardrailFormInput;
  onChange: (patch: SemanticThresholdPatch) => void;
};

function clampUnit(value: string) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return 0;

  return Math.min(1, Math.max(0, parsed));
}

/** Three bands on one bar; in RTL the bar grows from the right like the rest of the layout. */
function ThresholdBand({ thresholds, label }: { thresholds: SemanticThresholdInput; label: string }) {
  const review = clampUnit(thresholds.review);
  const block = Math.max(review, clampUnit(thresholds.block));

  return (
    <div role="img" aria-label={label} className="flex h-2 w-full overflow-hidden rounded-full bg-muted">
      <span className="h-full bg-success" style={{ width: `${review * 100}%` }} />
      <span className="h-full bg-warning" style={{ width: `${(block - review) * 100}%` }} />
      <span className="h-full flex-1 bg-destructive" />
    </div>
  );
}

type ThresholdRowProps = {
  id: string;
  label: string;
  thresholds: SemanticThresholdInput;
  onChange: (thresholds: SemanticThresholdInput) => void;
};

function ThresholdRow({ id, label, thresholds, onChange }: ThresholdRowProps) {
  const t = useTranslations("guardrails");
  const percent = useNumberFormatter({ style: "percent", maximumFractionDigits: 0 });
  const invalid = !validThresholds(thresholds);
  const bandLabel = t("wizard.behavior.semantic.bandLabel", {
    rule: label,
    review: percent.format(clampUnit(thresholds.review)),
    block: percent.format(clampUnit(thresholds.block)),
  });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium" dir="auto">{label}</span>
        <ThresholdBand thresholds={thresholds} label={bandLabel} />
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <AppNumberField
          id={`${id}-review`}
          label={t("wizard.behavior.semantic.review")}
          min={0}
          max={1}
          step={0.05}
          value={thresholds.review}
          error={invalid ? t("validation.semanticThresholds") : undefined}
          onChange={(event) => onChange({ ...thresholds, review: event.target.value })}
        />
        <AppNumberField
          id={`${id}-block`}
          label={t("wizard.behavior.semantic.block")}
          min={0}
          max={1}
          step={0.05}
          value={thresholds.block}
          onChange={(event) => onChange({ ...thresholds, block: event.target.value })}
        />
      </div>
    </div>
  );
}

function BandLegend() {
  const t = useTranslations("guardrails");
  const bands = [
    { key: "pass", className: "bg-success" },
    { key: "review", className: "bg-warning" },
    { key: "block", className: "bg-destructive" },
  ] as const;

  return (
    <ul className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
      {bands.map((band) => (
        <li key={band.key} className="flex items-center gap-1.5">
          <span aria-hidden="true" className={cn("size-2 rounded-full", band.className)} />
          {t(`wizard.behavior.semantic.bands.${band.key}`)}
        </li>
      ))}
    </ul>
  );
}

export function SemanticThresholds({ form, onChange }: SemanticThresholdsProps) {
  const t = useTranslations("guardrails");
  const headingID = useId();
  const pack = form.semanticSource === "pack";
  const rows = pack
    ? form.semanticPackRules.map((rule, index) => ({
        id: `pack-${rule.key}`,
        label: t(`wizard.config.semantic.packRules.${rule.key}.title`),
        thresholds: rule,
        update: (thresholds: SemanticThresholdInput) => onChange({ semanticPackRules: form.semanticPackRules.map((item, current) => (current === index ? { ...item, ...thresholds } : item)) }),
      }))
    : form.semanticRules.map((rule, index) => ({
        id: rule.id,
        label: rule.statement.trim() || rule.key,
        thresholds: rule,
        update: (thresholds: SemanticThresholdInput) => onChange({ semanticRules: form.semanticRules.map((item, current) => (current === index ? { ...item, ...thresholds } : item)) }),
      }));

  return (
    <section className="flex flex-col gap-5" aria-labelledby={headingID}>
      <div className="flex flex-col gap-1">
        <h3 id={headingID} className="text-sm font-semibold">{t("wizard.behavior.semantic.thresholdsTitle")}</h3>
        <p className="text-sm leading-6 text-muted-foreground">{t("wizard.behavior.semantic.thresholdsHint")}</p>
        <BandLegend />
      </div>

      {rows.map((row) => (
        <ThresholdRow key={row.id} id={row.id} label={row.label} thresholds={{ review: row.thresholds.review, block: row.thresholds.block }} onChange={row.update} />
      ))}

      <AppSwitchField
        id="semantic-finglish"
        label={t("wizard.behavior.semantic.finglishTitle")}
        description={t("wizard.behavior.semantic.finglishHint")}
        checked={form.semanticFinglish}
        onCheckedChange={(semanticFinglish) => onChange({ semanticFinglish })}
      />
      {form.semanticFinglish ? (
        <ThresholdRow
          id="semantic-finglish-thresholds"
          label={t("wizard.behavior.semantic.finglishRow")}
          thresholds={form.semanticFinglishThresholds}
          onChange={(semanticFinglishThresholds) => onChange({ semanticFinglishThresholds })}
        />
      ) : null}

      <AppSwitchField
        id="semantic-severity"
        label={t("wizard.behavior.semantic.severityTitle")}
        description={t("wizard.behavior.semantic.severityHint")}
        checked={form.semanticSeverity}
        onCheckedChange={(semanticSeverity) => onChange({ semanticSeverity })}
      />
      {form.semanticSeverity ? (
        <AppNumberField
          id="semantic-severity-block-at"
          label={t("wizard.behavior.semantic.severityBlockAt")}
          description={t("wizard.behavior.semantic.severityBlockAtHint", { max: MAX_SEMANTIC_SEVERITY })}
          fieldClassName="max-w-xs"
          min={0}
          max={MAX_SEMANTIC_SEVERITY}
          step={0.5}
          value={form.semanticSeverityBlockAt}
          onChange={(event) => onChange({ semanticSeverityBlockAt: event.target.value })}
        />
      ) : null}
    </section>
  );
}

/** A beta pack has not been through the evaluation gate, so enforcing it right away risks blocking good traffic. */
export function SemanticBetaPackNote({ form }: { form: GuardrailFormInput }) {
  const t = useTranslations("guardrails");
  if (form.semanticSource !== "pack" || !SEMANTIC_PACKS[form.semanticPack].beta || form.enforcement === "audit") return null;

  return (
    <p className="flex items-start gap-2 text-xs leading-5 text-warning-soft-foreground">
      <TriangleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
      {t("wizard.behavior.semantic.betaPackAudit")}
    </p>
  );
}

/** What happens when the judge cannot answer, which differs by enforcement for semantic policies. */
export function SemanticFailureNote({ enforcement }: { enforcement: GuardrailEnforcement }) {
  const t = useTranslations("guardrails");

  return <p className="text-xs leading-5 text-muted-foreground">{t(`wizard.behavior.semantic.failure.${enforcement}`)}</p>;
}
