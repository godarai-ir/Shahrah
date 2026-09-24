"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import { CircleAlert, CircleCheck, Info, Plus } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { Button } from "@/components/ui/button";
import { FieldDescription, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { useNumberFormatter } from "@/hooks/use-locale-format";
import { cn } from "@/lib/utils";

import { catalogEntry } from "../catalog";
import { JWT_VERIFIERS, lines, parseJSONOrEmpty } from "../form";
import type { GuardrailFormInput } from "../types";
import { SemanticPolicyConfig } from "./SemanticPolicyConfig";

type ConfigPatch = Partial<Omit<GuardrailFormInput, "type" | "name" | "slug">>;

type StepConfigProps = {
  form: GuardrailFormInput;
  dir: "rtl" | "ltr";
  headingRef?: React.Ref<HTMLHeadingElement>;
  onChange: (patch: ConfigPatch) => void;
};

const TERM_PRESETS = ["ignore previous instructions", "system prompt", "jailbreak"];

const PATTERN_PRESETS = [
  { key: "password", pattern: "(?i)password\\s*[:=]\\s*\\S+" },
  { key: "card", pattern: "\\b(?:\\d{4}[- ]?){3}\\d{4}\\b" },
  { key: "nationalId", pattern: "\\b\\d{10}\\b" },
  { key: "email", pattern: "[\\w.+-]+@[\\w-]+\\.[\\w.]+" },
] as const;

const BYTES_PER_MEGABYTE = 1024 * 1024;

function isValidJSON(value: string) {
  try {
    parseJSONOrEmpty(value);
    return true;
  } catch {
    return false;
  }
}

function PresetChips({ label, presets, onAdd, technical }: { label: string; presets: { key: string; label: string; value: string }[]; onAdd: (value: string) => void; technical?: boolean }) {
  const t = useTranslations("guardrails");

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-muted-foreground">{label}</span>
      {presets.map((preset) => (
        <Button
          key={preset.key}
          type="button"
          size="sm"
          variant="outline"
          aria-label={t("wizard.config.addPreset", { label: preset.label })}
          onClick={() => onAdd(preset.value)}
          className={cn("rounded-full border-dashed font-normal", technical && "font-mono")}
        >
          <Plus data-icon="inline-start" aria-hidden="true" />
          <span dir={technical ? "ltr" : undefined}>{preset.label}</span>
        </Button>
      ))}
    </div>
  );
}

function CountRange({ form, onChange }: Pick<StepConfigProps, "form" | "onChange">) {
  const t = useTranslations("guardrails");
  const hintID = useId();
  const unit = catalogEntry(form.type).unit ?? "word";

  return (
    <FieldSet className="gap-3">
      <FieldLegend variant="label" className="mb-0">{t("wizard.config.rangeLabel")}</FieldLegend>
      <FieldDescription id={hintID}>{t("wizard.config.rangeHint")}</FieldDescription>
      <div className="flex flex-wrap items-center gap-3 text-sm">
        <span>{t("wizard.config.rangeBetween")}</span>
        <Input
          type="number"
          inputMode="numeric"
          min={0}
          dir="ltr"
          aria-label={t("form.min")}
          aria-describedby={hintID}
          placeholder={t("form.min")}
          value={form.min}
          onChange={(event) => onChange({ min: event.target.value })}
          className="w-28 text-center tabular-nums"
        />
        <span>{t("wizard.config.rangeAnd")}</span>
        <Input
          type="number"
          inputMode="numeric"
          min={0}
          dir="ltr"
          aria-label={t("form.max")}
          aria-describedby={hintID}
          placeholder={t("form.max")}
          value={form.max}
          onChange={(event) => onChange({ max: event.target.value })}
          className="w-28 text-center tabular-nums"
        />
        <span>{t(`wizard.config.rangeUnits.${unit}`)}</span>
      </div>
    </FieldSet>
  );
}

function JwtFields({ form, dir, onChange }: Pick<StepConfigProps, "form" | "dir" | "onChange">) {
  const t = useTranslations("guardrails");
  const legendID = useId();
  const optional = (label: string) => t("wizard.config.optionalLabel", { label });

  return (
    <>
      <FieldSet className="gap-3">
        <FieldLegend id={legendID} variant="label" className="mb-0">{t("wizard.config.verifierLabel")}</FieldLegend>
        <OptionCardGroup
          aria-labelledby={legendID}
          dir={dir}
          columns={4}
          density="compact"
          value={form.jwtVerifier}
          onValueChange={(jwtVerifier) => onChange({ jwtVerifier })}
          options={JWT_VERIFIERS.map((verifier) => ({
            value: verifier,
            title: t(`wizard.config.verifiers.${verifier}`),
            description: t(`wizard.config.verifierHints.${verifier}`),
          }))}
        />
      </FieldSet>

      {form.jwtVerifier === "jwks" ? (
        <AppTextField id="guardrail-jwks-uri" label={t("form.jwksUri")} dir="ltr" type="url" value={form.jwksUri} placeholder="https://auth.example.com/.well-known/jwks.json" onChange={(event) => onChange({ jwksUri: event.target.value })} />
      ) : null}
      {form.jwtVerifier === "inline" ? (
        <AppCodeTextareaField id="guardrail-inline-jwks" label={t("form.inlineJwks")} className="min-h-32" value={form.inlineJwks} placeholder='{ "keys": [] }' onChange={(event) => onChange({ inlineJwks: event.target.value })} />
      ) : null}
      {form.jwtVerifier === "shared" ? (
        <AppTextField id="guardrail-shared-secret" label={t("form.sharedSecret")} dir="ltr" type="password" autoComplete="off" value={form.sharedSecret} onChange={(event) => onChange({ sharedSecret: event.target.value })} />
      ) : null}
      {form.jwtVerifier === "introspection" ? (
        <div className="grid gap-4 md:grid-cols-2">
          <AppTextField id="guardrail-introspection" label={t("form.introspectionUrl")} dir="ltr" type="url" value={form.introspectionUrl} placeholder="https://auth.example.com/introspect" onChange={(event) => onChange({ introspectionUrl: event.target.value })} />
          <AppTextField id="guardrail-introspection-auth" label={optional(t("form.introspectionAuthHeader"))} dir="ltr" value={form.introspectionAuthHeader} placeholder="Bearer ..." onChange={(event) => onChange({ introspectionAuthHeader: event.target.value })} />
        </div>
      ) : null}

      <Separator />
      <div className="grid gap-4 md:grid-cols-3">
        <AppTextField id="guardrail-issuer" label={optional(t("form.issuer"))} dir="ltr" value={form.issuer} onChange={(event) => onChange({ issuer: event.target.value })} />
        <AppTextField id="guardrail-audience" label={optional(t("form.audience"))} dir="ltr" value={form.audience} onChange={(event) => onChange({ audience: event.target.value })} />
        <AppTextField id="guardrail-token-regex" label={optional(t("form.tokenRegex"))} dir="ltr" className="font-mono" value={form.tokenRegex} placeholder="Bearer (.+)" onChange={(event) => onChange({ tokenRegex: event.target.value })} />
      </div>
    </>
  );
}

function ConfigFields({ form, dir, onChange }: Omit<StepConfigProps, "headingRef">) {
  const t = useTranslations("guardrails");
  const legendID = useId();
  const megabytes = useNumberFormatter({ minimumFractionDigits: 1, maximumFractionDigits: 1 });
  const addLine = (field: "terms" | "patterns", value: string) => {
    const current = lines(form[field]);
    if (current.includes(value)) return;
    onChange({ [field]: [...current, value].join("\n") });
  };

  switch (form.type) {
    case "keyword_blocklist":
      return (
        <>
          <AppTextareaField id="guardrail-terms" label={t("form.terms")} description={t("wizard.config.termsHint")} dir="ltr" className="min-h-32 font-mono" value={form.terms} placeholder={"ignore previous instructions\ndrop table"} onChange={(event) => onChange({ terms: event.target.value })} />
          <PresetChips technical label={t("wizard.config.quickAdd")} presets={TERM_PRESETS.map((term) => ({ key: term, label: term, value: term }))} onAdd={(value) => addLine("terms", value)} />
        </>
      );
    case "regex":
      return (
        <>
          <AppTextareaField id="guardrail-patterns" label={t("form.patterns")} description={t("wizard.config.patternsHint")} dir="ltr" className="min-h-28 font-mono" value={form.patterns} placeholder="(?i)password\s*[:=]\s*\S+" onChange={(event) => onChange({ patterns: event.target.value })} />
          <PresetChips label={t("wizard.config.presets")} presets={PATTERN_PRESETS.map((preset) => ({ key: preset.key, label: t(`wizard.config.patternPresets.${preset.key}`), value: preset.pattern }))} onAdd={(value) => addLine("patterns", value)} />
          <AppTextField id="guardrail-redaction" label={t("form.redactionText")} description={t("wizard.config.redactionHint")} dir="ltr" className="max-w-xs font-mono" value={form.redactionText} onChange={(event) => onChange({ redactionText: event.target.value })} />
        </>
      );
    case "regex_replace":
      return (
        <>
          <AppTextField id="guardrail-pattern" label={t("form.pattern")} description={t("wizard.config.patternHint")} dir="ltr" className="font-mono" value={form.pattern} placeholder="(?i)password\s*[:=]\s*\S+" onChange={(event) => onChange({ pattern: event.target.value })} />
          <AppTextField id="guardrail-replacement" label={t("form.replacement")} description={t("wizard.config.replacementHint")} dir="ltr" className="max-w-xs font-mono" value={form.replacement} onChange={(event) => onChange({ replacement: event.target.value })} />
        </>
      );
    case "secrets_detection":
    case "basic_pii":
      return (
        <AppTextField id="guardrail-redaction" label={t("form.redactionText")} description={t("wizard.config.redactionHint")} dir="ltr" className="max-w-xs font-mono" value={form.redactionText} onChange={(event) => onChange({ redactionText: event.target.value })} />
      );
    case "sentence_count":
    case "word_count":
    case "character_count":
      return <CountRange form={form} onChange={onChange} />;
    case "json_schema": {
      const valid = isValidJSON(form.schema);
      const StatusIcon = valid ? CircleCheck : CircleAlert;

      return (
        <>
          <AppCodeTextareaField id="guardrail-schema" label={t("form.schema")} description={t("wizard.config.schemaHint")} className="min-h-48" value={form.schema} onChange={(event) => onChange({ schema: event.target.value })} />
          <p aria-live="polite" className={cn("flex items-center gap-2 text-xs", valid ? "text-success-soft-foreground" : "text-destructive-soft-foreground")}>
            <StatusIcon className="size-4" aria-hidden="true" />
            {valid ? t("wizard.config.schemaValid") : t("wizard.config.schemaInvalid")}
          </p>
        </>
      );
    }
    case "json_keys":
      return (
        <>
          <AppTextareaField id="guardrail-keys" label={t("form.keys")} description={t("wizard.config.keysHint")} dir="ltr" className="min-h-24 font-mono" value={form.keys} placeholder={"answer\nconfidence"} onChange={(event) => onChange({ keys: event.target.value })} />
          <FieldSet className="gap-3">
            <FieldLegend id={legendID} variant="label" className="mb-0">{t("form.keyMode")}</FieldLegend>
            <OptionCardGroup
              aria-labelledby={legendID}
              dir={dir}
              columns={3}
              density="compact"
              value={form.keyMode}
              onValueChange={(keyMode) => onChange({ keyMode })}
              options={(["all", "any", "none"] as const).map((mode) => ({ value: mode, title: t(`keyModes.${mode}`), description: t(`wizard.config.keyModeHints.${mode}`) }))}
            />
          </FieldSet>
        </>
      );
    case "contains_code":
      return (
        <>
          <FieldSet className="gap-3">
            <FieldLegend id={legendID} variant="label" className="mb-0">{t("form.codeMode")}</FieldLegend>
            <OptionCardGroup
              aria-labelledby={legendID}
              dir={dir}
              density="compact"
              value={form.codeMode}
              onValueChange={(codeMode) => onChange({ codeMode })}
              options={(["must_not_contain", "must_contain"] as const).map((mode) => ({ value: mode, title: t(`codeModes.${mode}`), description: t(`wizard.config.codeModeHints.${mode}`) }))}
            />
          </FieldSet>
          <AppTextareaField id="guardrail-languages" label={t("wizard.config.optionalLabel", { label: t("form.languages") })} description={t("wizard.config.languagesHint")} dir="ltr" className="min-h-20 font-mono" value={form.languages} placeholder={"sql\npython"} onChange={(event) => onChange({ languages: event.target.value })} />
        </>
      );
    case "jwt_validation":
      return <JwtFields form={form} dir={dir} onChange={onChange} />;
    case "inline_image_urls": {
      const bytes = Number(form.maxBytes);

      return (
        <>
          <AppTextareaField id="guardrail-allowed-hosts" label={t("form.allowedHosts")} description={t("wizard.config.allowedHostsHint")} dir="ltr" className="min-h-24 font-mono" value={form.allowedHosts} placeholder={"cdn.example.com\nimages.example.com"} onChange={(event) => onChange({ allowedHosts: event.target.value })} />
          <div className="grid gap-4 md:grid-cols-2">
            <AppNumberField
              id="guardrail-max-bytes"
              label={t("form.maxBytes")}
              min={0}
              description={bytes > 0 ? t("wizard.config.maxBytesHuman", { size: megabytes.format(bytes / BYTES_PER_MEGABYTE) }) : undefined}
              value={form.maxBytes}
              onChange={(event) => onChange({ maxBytes: event.target.value })}
            />
            <AppNumberField id="guardrail-timeout" label={t("form.timeoutMs")} min={0} value={form.timeoutMs} onChange={(event) => onChange({ timeoutMs: event.target.value })} />
          </div>
          <AppSwitchField id="guardrail-private-networks" label={t("form.allowPrivateNetworks")} description={t("wizard.config.privateNetworksHint")} checked={form.allowPrivateNetworks} onCheckedChange={(allowPrivateNetworks) => onChange({ allowPrivateNetworks })} />
        </>
      );
    }
    case "semantic_policy":
      return <SemanticPolicyConfig form={form} dir={dir} onChange={onChange} />;
    case "not_null":
      return (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <Info className="size-4 shrink-0" aria-hidden="true" />
          {t("wizard.config.noConfig")}
        </p>
      );
  }
}

export function StepConfig({ form, dir, headingRef, onChange }: StepConfigProps) {
  const t = useTranslations("guardrails");
  const Icon = catalogEntry(form.type).icon;
  const typeName = t(`types.${form.type}`);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-start gap-3">
        <span aria-hidden="true" className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary [&_svg]:size-5">
          <Icon />
        </span>
        <div className="flex min-w-0 flex-col gap-1">
          <h2 ref={headingRef} tabIndex={-1} className="text-base font-semibold outline-none md:text-lg">{t("wizard.config.heading", { type: typeName })}</h2>
          <p className="text-sm leading-6 text-muted-foreground">{t(`typeDescriptions.${form.type}`)}</p>
        </div>
      </div>
      <AppCard contentClassName="flex flex-col gap-5">
        <ConfigFields form={form} dir={dir} onChange={onChange} />
      </AppCard>
    </div>
  );
}
