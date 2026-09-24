"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import { ArrowDownToLine, ArrowUpFromLine, Info, PencilRuler, Plus, ShieldCheck, Trash2, TriangleAlert } from "lucide-react";

import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { Button } from "@/components/ui/button";
import { FieldDescription, FieldLegend, FieldSet } from "@/components/ui/field";
import { Separator } from "@/components/ui/separator";

import { MAX_SEMANTIC_RULES, newSemanticRule, nextRuleKey, SEMANTIC_PACK_RULE_KEYS, statementWarnings } from "../semantic";
import type { GuardrailFormInput, SemanticRuleInput, SemanticSource, SemanticTarget } from "../types";

type SemanticPatch = Partial<Pick<GuardrailFormInput, "semanticSource" | "semanticTarget" | "semanticRules">>;

type SemanticPolicyConfigProps = {
  form: GuardrailFormInput;
  dir: "rtl" | "ltr";
  onChange: (patch: SemanticPatch) => void;
};

const SOURCES: SemanticSource[] = ["pack", "custom"];
const SOURCE_ICONS = { pack: ShieldCheck, custom: PencilRuler } as const;
const TARGETS: SemanticTarget[] = ["input", "output"];
const TARGET_ICONS = { input: ArrowDownToLine, output: ArrowUpFromLine } as const;

function PackRules() {
  const t = useTranslations("guardrails");

  return (
    <div className="flex flex-col gap-3">
      <p className="text-sm leading-6 text-muted-foreground">{t("wizard.config.semantic.packIntro")}</p>
      <ul className="flex flex-col gap-3">
        {SEMANTIC_PACK_RULE_KEYS.map((key) => (
          <li key={key} className="flex items-start gap-3">
            <ShieldCheck className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden="true" />
            <span className="flex min-w-0 flex-col gap-0.5">
              <span className="text-sm font-medium">{t(`wizard.config.semantic.packRules.${key}.title`)}</span>
              <span className="text-xs leading-5 text-muted-foreground">{t(`wizard.config.semantic.packRules.${key}.description`)}</span>
            </span>
          </li>
        ))}
      </ul>
    </div>
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
  const targetID = useId();
  const rules = form.semanticRules;
  const atLimit = rules.length >= MAX_SEMANTIC_RULES;
  const updateRule = (index: number, patch: Partial<SemanticRuleInput>) => {
    onChange({ semanticRules: rules.map((rule, current) => (current === index ? { ...rule, ...patch } : rule)) });
  };

  return (
    <>
      <FieldSet className="gap-3">
        <FieldLegend id={targetID} variant="label" className="mb-0">{t("wizard.config.semantic.targetLabel")}</FieldLegend>
        <OptionCardGroup
          aria-labelledby={targetID}
          dir={dir}
          density="compact"
          value={form.semanticTarget}
          onValueChange={(semanticTarget) => onChange({ semanticTarget })}
          options={TARGETS.map((value) => ({
            value,
            icon: TARGET_ICONS[value],
            title: t(`wizard.config.semantic.targets.${value}`),
            description: t(`wizard.config.semantic.targetHints.${value}`),
          }))}
        />
      </FieldSet>

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
          onValueChange={(semanticSource) => onChange({ semanticSource })}
          options={SOURCES.map((value) => ({
            value,
            icon: SOURCE_ICONS[value],
            title: t(`wizard.config.semantic.sources.${value}`),
            description: t(`wizard.config.semantic.sourceHints.${value}`),
            tip: value === "pack" ? t("wizard.config.semantic.packTip") : undefined,
          }))}
        />
      </FieldSet>

      {form.semanticSource === "pack" ? <PackRules /> : <CustomRules form={form} dir={dir} onChange={onChange} />}

      <p className="flex items-start gap-2 text-xs leading-5 text-muted-foreground">
        <Info className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
        {t("wizard.config.semantic.privacyNote")}
      </p>
    </>
  );
}
