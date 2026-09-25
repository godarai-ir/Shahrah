"use client";

import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppTextareaField } from "@/components/app-ui/AppTextField";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { Button } from "@/components/ui/button";
import { useNumberFormatter } from "@/hooks/use-locale-format";
import { cn } from "@/lib/utils";

import { catalogEntry } from "../catalog";
import { isValidSlug, lines, type WizardStep } from "../form";
import { SEMANTIC_PACKS, semanticCostPerThousand, semanticRuleCount } from "../semantic";
import type { SentenceFormat } from "../sentence";
import type { GuardrailFormInput } from "../types";
import type { WizardMode } from "./state";

type StepReviewProps = {
  form: GuardrailFormInput;
  mode: WizardMode;
  sentence: string;
  existingSlugs: readonly string[];
  format: SentenceFormat;
  headingRef?: React.Ref<HTMLHeadingElement>;
  onNameChange: (name: string) => void;
  onSlugChange: (slug: string) => void;
  onDescriptionChange: (description: string) => void;
  onEditStep: (step: WizardStep) => void;
};

/** A compact, readable summary of the type-specific settings for the review table. */
function configSummary(form: GuardrailFormInput, { t, formatNumber }: SentenceFormat): { value: string; technical: boolean } {
  const separator = t("sentence.listSeparator");
  const empty = t("wizard.review.configSummary.empty");
  const joined = (value: string, glue = "\n") => lines(value).join(glue) || empty;

  switch (form.type) {
    case "keyword_blocklist":
      return { value: joined(form.terms), technical: true };
    case "regex":
      return { value: joined(form.patterns), technical: true };
    case "regex_replace":
      return { value: `${form.pattern.trim() || empty} → ${form.replacement}`, technical: true };
    case "secrets_detection":
    case "basic_pii":
      return { value: t("wizard.review.configSummary.redaction", { text: form.redactionText.trim() || "[***]" }), technical: false };
    case "sentence_count":
    case "word_count":
    case "character_count": {
      const unbounded = t("wizard.review.configSummary.unbounded");
      const bound = (value: string) => (value.trim() && Number.isFinite(Number(value)) ? formatNumber(Number(value)) : unbounded);
      const unit = t(`wizard.config.rangeUnits.${catalogEntry(form.type).unit}`);

      return { value: t("wizard.review.configSummary.range", { min: bound(form.min), max: bound(form.max), unit }), technical: false };
    }
    case "json_schema":
      return { value: form.schema.trim() || empty, technical: true };
    case "json_keys":
      return { value: `${joined(form.keys, separator)} · ${t(`keyModes.${form.keyMode}`)}`, technical: false };
    case "contains_code": {
      const languages = lines(form.languages);
      const mode = t(`codeModes.${form.codeMode}`);

      return { value: languages.length > 0 ? `${mode} · ${languages.join(separator)}` : mode, technical: false };
    }
    case "jwt_validation": {
      const method = t(`wizard.config.verifiers.${form.jwtVerifier}`);
      const address = form.jwtVerifier === "jwks" ? form.jwksUri.trim() : form.jwtVerifier === "introspection" ? form.introspectionUrl.trim() : "";

      return { value: address ? `${method} · ${address}` : method, technical: false };
    }
    case "inline_image_urls": {
      const hosts = lines(form.allowedHosts).join(separator) || t("wizard.review.configSummary.allHosts");
      const timeout = Number(form.timeoutMs);
      const timeoutText = Number.isFinite(timeout) && form.timeoutMs.trim() ? ` · ${t("wizard.review.configSummary.timeout", { ms: formatNumber(timeout) })}` : "";

      return { value: `${hosts}${timeoutText}`, technical: false };
    }
    case "not_null":
      return { value: t("wizard.review.configSummary.none"), technical: false };
    case "document_confidentiality": {
      const level = t(`wizard.config.confidentiality.levels.${form.confidentialityLevel}.title`);
      const markings = lines(form.customMarkings);
      const custom = markings.length > 0 ? ` · ${t("wizard.review.configSummary.customMarkings", { markings: markings.join(separator) })}` : "";

      return { value: `${t("wizard.review.configSummary.confidentiality", { level })}${custom}`, technical: false };
    }
    case "semantic_policy": {
      const count = semanticRuleCount(form);
      const target = t(`wizard.config.semantic.targets.${form.semanticTarget}`);
      if (form.semanticSource === "pack") {
        const pack = t(`wizard.config.semantic.packs.${SEMANTIC_PACKS[form.semanticPack].family}.title`);

        return { value: t("wizard.review.configSummary.semanticPack", { n: count, count: formatNumber(count), pack, target }), technical: false };
      }
      const keys = form.semanticRules.map((rule) => rule.key.trim()).join(separator);

      return { value: t("wizard.review.configSummary.semanticCustom", { n: count, count: formatNumber(count), target, keys }), technical: false };
    }
  }
}

export function StepReview({ form, mode, sentence, existingSlugs, format, headingRef, onNameChange, onSlugChange, onDescriptionChange, onEditStep }: StepReviewProps) {
  const t = useTranslations("guardrails");
  const slug = form.slug.trim();
  const slugTaken = mode === "create" && slug !== "" && existingSlugs.includes(slug);
  const slugError = mode === "create" && slug !== "" && !isValidSlug(slug) ? t("validation.slug") : slugTaken ? t("messages.duplicate") : undefined;
  const currency = useNumberFormatter({ style: "currency", currency: "USD", maximumFractionDigits: 3 });
  const config = configSummary(form, format);
  const rows: { key: "type" | "config" | "behavior" | "status" | "cost"; value: string; step: WizardStep; technical?: boolean }[] = [
    { key: "type", value: t(`types.${form.type}`), step: 0 },
    { key: "config", value: config.value, step: 1, technical: config.technical },
    { key: "behavior", value: sentence, step: 2 },
    { key: "status", value: form.enabled ? t("status.enabled") : t("status.disabled"), step: 2 },
  ];
  if (form.type === "semantic_policy") {
    rows.push({ key: "cost", value: t("wizard.review.costValue", { cost: currency.format(semanticCostPerThousand(form)) }), step: 1 });
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h2 ref={headingRef} tabIndex={-1} className="text-base font-semibold outline-none md:text-lg">{t("wizard.review.title")}</h2>
        <p className="text-sm leading-6 text-muted-foreground">{t("wizard.review.slugLocked")}</p>
      </div>

      <AppCard contentClassName="flex flex-col gap-5">
        <div className="grid gap-4 md:grid-cols-2">
          <DisplayNameField id="guardrail-name" label={t("form.name")} value={form.name} placeholder={t("wizard.review.namePlaceholder")} onChange={(event) => onNameChange(event.target.value)} />
          <IdentifierField
            id="guardrail-slug"
            label={t("form.identifier")}
            pattern="[A-Za-z0-9][A-Za-z0-9_-]{1,95}"
            minLength={2}
            maxLength={96}
            value={form.slug}
            disabled={mode === "edit"}
            error={slugError}
            description={slugError ? undefined : mode === "edit" ? t("wizard.review.slugLocked") : t("wizard.review.slugHint")}
            onChange={(event) => onSlugChange(event.target.value)}
          />
        </div>
        <AppTextareaField
          id="guardrail-description"
          label={t("wizard.config.optionalLabel", { label: t("form.notes") })}
          value={form.description}
          placeholder={t("wizard.review.descriptionPlaceholder")}
          onChange={(event) => onDescriptionChange(event.target.value)}
        />
      </AppCard>

      <AppCard title={t("wizard.review.heading")} contentClassName="p-0 md:p-0">
        <dl className="flex flex-col">
          {rows.map((row) => (
            <div key={row.key} className="flex items-start gap-3 border-t border-border px-4 py-3 md:px-5">
              <dt className="w-24 shrink-0 text-sm text-muted-foreground">{t(`wizard.review.rows.${row.key}`)}</dt>
              <dd
                dir={row.technical ? "ltr" : undefined}
                className={cn("min-w-0 flex-1 whitespace-pre-wrap break-words text-start text-sm leading-6", row.technical && "font-mono text-xs leading-6")}
              >
                {row.value}
              </dd>
              {mode === "edit" && row.step === 0 ? null : (
                <Button
                  type="button"
                  variant="link"
                  size="sm"
                  className="h-auto min-w-0 shrink-0 p-0"
                  aria-label={t("wizard.review.editRow", { row: t(`wizard.review.rows.${row.key}`) })}
                  onClick={() => onEditStep(row.step)}
                >
                  {t("wizard.review.edit")}
                </Button>
              )}
            </div>
          ))}
        </dl>
      </AppCard>
    </div>
  );
}
