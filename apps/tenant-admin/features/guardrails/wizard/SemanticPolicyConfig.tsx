"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import {
  ArrowDownToLine,
  ArrowUpFromLine,
  Info,
  Landmark,
  PencilRuler,
  Plus,
  ShieldCheck,
  Stethoscope,
  Trash2,
  TriangleAlert,
  UsersRound,
} from "lucide-react";

import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { Button } from "@/components/ui/button";
import { FieldDescription, FieldLegend, FieldSet } from "@/components/ui/field";
import { Separator } from "@/components/ui/separator";

import {
  MAX_SEMANTIC_RULES,
  newSemanticRule,
  nextRuleKey,
  packRuleDefinitions,
  packSelection,
  packTargets,
  SEMANTIC_PACK_IDS,
  SEMANTIC_PACKS,
  statementWarnings,
} from "../semantic";
import type { GuardrailFormInput, SemanticPackId, SemanticRuleInput, SemanticSource, SemanticTarget } from "../types";

type SemanticPatch = Partial<Pick<GuardrailFormInput, "semanticSource" | "semanticPack" | "semanticTarget" | "semanticPackRules" | "semanticRules">>;

type SemanticPolicyConfigProps = {
  form: GuardrailFormInput;
  dir: "rtl" | "ltr";
  onChange: (patch: SemanticPatch) => void;
};

const SOURCES: SemanticSource[] = ["pack", "custom"];
const SOURCE_ICONS = { pack: ShieldCheck, custom: PencilRuler } as const;
const TARGETS: SemanticTarget[] = ["input", "output"];
const TARGET_ICONS = { input: ArrowDownToLine, output: ArrowUpFromLine } as const;
const PACK_ICONS = { security_baseline: ShieldCheck, banking: Landmark, healthcare: Stethoscope, hr: UsersRound } as const;

type TargetPickerProps = {
  value: SemanticTarget;
  dir: "rtl" | "ltr";
  /** Hooks the current source has no rules for; they stay visible but cannot be picked. */
  unavailable?: readonly SemanticTarget[];
  lockedNote?: string;
  onChange: (target: SemanticTarget) => void;
};

function TargetPicker({ value, dir, unavailable = [], lockedNote, onChange }: TargetPickerProps) {
  const t = useTranslations("guardrails");
  const targetID = useId();

  return (
    <FieldSet className="gap-3">
      <FieldLegend id={targetID} variant="label" className="mb-0">{t("wizard.config.semantic.targetLabel")}</FieldLegend>
      {lockedNote ? <FieldDescription>{lockedNote}</FieldDescription> : null}
      <OptionCardGroup
        aria-labelledby={targetID}
        dir={dir}
        density="compact"
        value={value}
        onValueChange={onChange}
        options={TARGETS.map((target) => ({
          value: target,
          icon: TARGET_ICONS[target],
          title: t(`wizard.config.semantic.targets.${target}`),
          description: t(`wizard.config.semantic.targetHints.${target}`),
          disabled: unavailable.includes(target),
        }))}
      />
    </FieldSet>
  );
}

function PackConfig({ form, dir, onChange }: SemanticPolicyConfigProps) {
  const t = useTranslations("guardrails");
  const packID = useId();
  const pack = SEMANTIC_PACKS[form.semanticPack];
  const targets = packTargets(form.semanticPack);
  const unavailable = TARGETS.filter((target) => !targets.includes(target));
  const rules = packRuleDefinitions(form.semanticPack, form.semanticTarget);

  return (
    <>
      <FieldSet className="gap-3">
        <FieldLegend id={packID} variant="label" className="mb-0">{t("wizard.config.semantic.packLabel")}</FieldLegend>
        <OptionCardGroup
          aria-labelledby={packID}
          dir={dir}
          density="compact"
          value={form.semanticPack}
          onValueChange={(value: SemanticPackId) => onChange(packSelection(value, form.semanticTarget))}
          options={SEMANTIC_PACK_IDS.map((value) => {
            const family = SEMANTIC_PACKS[value].family;

            return {
              value,
              icon: PACK_ICONS[family],
              title: t(`wizard.config.semantic.packs.${family}.title`),
              description: t(`wizard.config.semantic.packs.${family}.description`),
              tip: t(SEMANTIC_PACKS[value].beta ? "wizard.config.semantic.packBetaTip" : "wizard.config.semantic.packTestedTip"),
            };
          })}
        />
      </FieldSet>

      <TargetPicker
        value={form.semanticTarget}
        dir={dir}
        unavailable={unavailable}
        lockedNote={unavailable.length > 0 ? t("wizard.config.semantic.packTargetLocked") : undefined}
        onChange={(target) => onChange(packSelection(form.semanticPack, target))}
      />

      <div className="flex flex-col gap-3">
        <p className="text-sm leading-6 text-muted-foreground">{t(`wizard.config.semantic.packIntro.${form.semanticTarget}`)}</p>
        <ul className="flex flex-col gap-3">
          {rules.map((rule) => (
            <li key={rule.key} className="flex items-start gap-3">
              <ShieldCheck className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden="true" />
              <span className="flex min-w-0 flex-col gap-0.5">
                <span className="text-sm font-medium">{t(`wizard.config.semantic.packRules.${rule.key}.title`)}</span>
                <span className="text-xs leading-5 text-muted-foreground">{t(`wizard.config.semantic.packRules.${rule.key}.description`)}</span>
                {rule.reviewOnly ? <span className="text-xs leading-5 text-warning-soft-foreground">{t("wizard.config.semantic.packRuleReviewOnly")}</span> : null}
              </span>
            </li>
          ))}
        </ul>
        {pack.beta ? (
          <p className="flex items-start gap-2 text-xs leading-5 text-warning-soft-foreground">
            <TriangleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
            {t("wizard.config.semantic.packBetaNote")}
          </p>
        ) : null}
      </div>
    </>
  );
}

type RuleEditorProps = {
  rule: SemanticRuleInput;
  index: number;
  target: SemanticTarget;
  removable: boolean;
  onChange: (patch: Partial<SemanticRuleInput>) => void;
  onRemove: () => void;
};

function RuleEditor({ rule, index, target, removable, onChange, onRemove }: RuleEditorProps) {
  const t = useTranslations("guardrails");
  const legendID = useId();
  const warnings = statementWarnings(rule.statement);
  const number = index + 1;
  const optional = (label: string) => t("wizard.config.optionalLabel", { label });

  return (
    <FieldSet aria-labelledby={legendID} className="gap-4">
      <div className="flex items-center justify-between gap-3">
        <FieldLegend id={legendID} variant="label" className="mb-0">{t("wizard.config.semantic.ruleLegend", { n: number })}</FieldLegend>
        {removable ? (
          <Button type="button" variant="ghost" size="sm" onClick={onRemove} aria-label={t("wizard.config.semantic.removeRule", { n: number })}>
            <Trash2 data-icon="inline-start" aria-hidden="true" />
            {t("wizard.config.semantic.remove")}
          </Button>
        ) : null}
      </div>

      <AppTextareaField
        id={`${rule.id}-statement`}
        label={t("wizard.config.semantic.statement")}
        description={t("wizard.config.semantic.statementHint")}
        dir="auto"
        rows={2}
        className="min-h-16"
        value={rule.statement}
        placeholder={t(`wizard.config.semantic.statementPlaceholder.${target}`)}
        onChange={(event) => onChange({ statement: event.target.value })}
      />
      {warnings.length > 0 ? (
        <ul aria-live="polite" className="flex flex-col gap-1">
          {warnings.map((warning) => (
            <li key={warning} className="flex items-start gap-2 text-xs leading-5 text-warning-soft-foreground">
              <TriangleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
              {t(`wizard.config.semantic.warnings.${warning}`)}
            </li>
          ))}
        </ul>
      ) : null}

      <div className="grid gap-4 md:grid-cols-2">
        <AppTextareaField id={`${rule.id}-true`} label={optional(t("wizard.config.semantic.trueWhen"))} dir="auto" rows={2} className="min-h-16" value={rule.trueWhen} placeholder={t("wizard.config.semantic.trueWhenPlaceholder")} onChange={(event) => onChange({ trueWhen: event.target.value })} />
        <AppTextareaField id={`${rule.id}-false`} label={optional(t("wizard.config.semantic.falseWhen"))} dir="auto" rows={2} className="min-h-16" value={rule.falseWhen} placeholder={t("wizard.config.semantic.falseWhenPlaceholder")} onChange={(event) => onChange({ falseWhen: event.target.value })} />
        <AppTextareaField id={`${rule.id}-examples-true`} label={optional(t("wizard.config.semantic.examplesTrue"))} description={t("wizard.config.semantic.examplesHint")} dir="auto" rows={3} className="min-h-20" value={rule.examplesTrue} onChange={(event) => onChange({ examplesTrue: event.target.value })} />
        <AppTextareaField id={`${rule.id}-examples-false`} label={optional(t("wizard.config.semantic.examplesFalse"))} description={t("wizard.config.semantic.examplesHint")} dir="auto" rows={3} className="min-h-20" value={rule.examplesFalse} onChange={(event) => onChange({ examplesFalse: event.target.value })} />
      </div>

      <AppTextField id={`${rule.id}-key`} label={t("wizard.config.semantic.key")} description={t("wizard.config.semantic.keyHint")} dir="ltr" fieldClassName="max-w-xs" className="font-mono" value={rule.key} onChange={(event) => onChange({ key: event.target.value })} />
    </FieldSet>
  );
}

function CustomRules({ form, dir, onChange }: SemanticPolicyConfigProps) {
  const t = useTranslations("guardrails");
  const rules = form.semanticRules;
  const atLimit = rules.length >= MAX_SEMANTIC_RULES;
  const updateRule = (index: number, patch: Partial<SemanticRuleInput>) => {
    onChange({ semanticRules: rules.map((rule, current) => (current === index ? { ...rule, ...patch } : rule)) });
  };

  return (
    <>
      <TargetPicker value={form.semanticTarget} dir={dir} onChange={(semanticTarget) => onChange({ semanticTarget })} />

      <p className="flex items-start gap-2 text-xs leading-5 text-muted-foreground">
        <Info className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
        {t("wizard.config.semantic.languageTip")}
      </p>

      {rules.map((rule, index) => (
        <div key={rule.id} className="flex flex-col gap-5">
          <Separator />
          <RuleEditor
            rule={rule}
            index={index}
            target={form.semanticTarget}
            removable={rules.length > 1}
            onChange={(patch) => updateRule(index, patch)}
            onRemove={() => onChange({ semanticRules: rules.filter((_, current) => current !== index) })}
          />
        </div>
      ))}

      <div className="flex flex-wrap items-center gap-3">
        <Button type="button" variant="outline" size="sm" disabled={atLimit} onClick={() => onChange({ semanticRules: [...rules, newSemanticRule(nextRuleKey(rules))] })}>
          <Plus data-icon="inline-start" aria-hidden="true" />
          {t("wizard.config.semantic.addRule")}
        </Button>
        {atLimit ? <FieldDescription>{t("wizard.config.semantic.maxRules", { max: MAX_SEMANTIC_RULES })}</FieldDescription> : null}
      </div>
    </>
  );
}

export function SemanticPolicyConfig({ form, dir, onChange }: SemanticPolicyConfigProps) {
  const t = useTranslations("guardrails");
  const sourceID = useId();

  return (
    <>
      <FieldSet className="gap-3">
        <FieldLegend id={sourceID} variant="label" className="mb-0">{t("wizard.config.semantic.sourceLabel")}</FieldLegend>
        <OptionCardGroup
          aria-labelledby={sourceID}
          dir={dir}
          density="compact"
          value={form.semanticSource}
          onValueChange={(semanticSource) =>
            onChange(semanticSource === "pack" ? { semanticSource, ...packSelection(form.semanticPack, form.semanticTarget) } : { semanticSource })
          }
          options={SOURCES.map((value) => ({
            value,
            icon: SOURCE_ICONS[value],
            title: t(`wizard.config.semantic.sources.${value}`),
            description: t(`wizard.config.semantic.sourceHints.${value}`),
            tip: value === "pack" ? t("wizard.config.semantic.packTip") : undefined,
          }))}
        />
      </FieldSet>

      {form.semanticSource === "pack" ? <PackConfig form={form} dir={dir} onChange={onChange} /> : <CustomRules form={form} dir={dir} onChange={onChange} />}

      <p className="flex items-start gap-2 text-xs leading-5 text-muted-foreground">
        <Info className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
        {t("wizard.config.semantic.privacyNote")}
      </p>
    </>
  );
}
