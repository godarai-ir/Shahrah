"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import { Eraser, Flag, ListChecks, Lock, OctagonMinus, OctagonX } from "lucide-react";

import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";

import { catalogEntry, operationForType } from "../catalog";
import type { GuardrailEnforcement, GuardrailFormInput, GuardrailOperation } from "../types";
import { SemanticFailureNote, SemanticThresholds } from "./SemanticThresholds";
import type { WizardMode } from "./state";

type StepBehaviorProps = {
  form: GuardrailFormInput;
  mode: WizardMode;
  dir: "rtl" | "ltr";
  headingRef?: React.Ref<HTMLHeadingElement>;
  onChange: (patch: Partial<Omit<GuardrailFormInput, "type" | "name" | "slug">>) => void;
};

const OPERATION_ICONS = { validate: ListChecks, mutate: Eraser } as const;
const ENFORCEMENT_ICONS = { enforce: OctagonX, audit: Flag, enforce_ignore_error: OctagonMinus } as const;
const ENFORCEMENTS: GuardrailEnforcement[] = ["enforce", "audit", "enforce_ignore_error"];
const OPERATIONS: GuardrailOperation[] = ["validate", "mutate"];

export function StepBehavior({ form, mode, dir, headingRef, onChange }: StepBehaviorProps) {
  const t = useTranslations("guardrails");
  const operationID = useId();
  const enforcementID = useId();
  const locked = catalogEntry(form.type).ops !== "both";
  const operation = operationForType(form.type, form.operation);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h2 ref={headingRef} tabIndex={-1} className="text-base font-semibold outline-none md:text-lg">{t("wizard.behavior.heading")}</h2>
        <p className="text-sm leading-6 text-muted-foreground">{t("wizard.behavior.description")}</p>
      </div>

      <section className="flex flex-col gap-3" aria-labelledby={operationID}>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <h3 id={operationID} className="text-sm font-semibold">{t("form.operation")}</h3>
          {locked ? (
            <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
              <Lock className="size-3" aria-hidden="true" />
              {t("wizard.behavior.opLocked", { type: t(`types.${form.type}`), operation: t(`operations.${operation}`) })}
            </span>
          ) : null}
        </div>
        <OptionCardGroup
          aria-labelledby={operationID}
          dir={dir}
          value={operation}
          onValueChange={(next) => onChange({ operation: next })}
          options={OPERATIONS.map((value) => ({
            value,
            icon: OPERATION_ICONS[value],
            title: t(`operations.${value}`),
            description: t(`operationHints.${value}`),
            disabled: locked && value !== operation,
          }))}
        />
      </section>

      <section className="flex flex-col gap-3" aria-labelledby={enforcementID}>
        <h3 id={enforcementID} className="text-sm font-semibold">{t("form.enforcement")}</h3>
        <OptionCardGroup
          aria-labelledby={enforcementID}
          dir={dir}
          columns={3}
          value={form.enforcement}
          onValueChange={(enforcement) => onChange({ enforcement })}
          options={ENFORCEMENTS.map((value) => ({
            value,
            icon: ENFORCEMENT_ICONS[value],
            title: t(`enforcement.${value}`),
            description: t(`enforcementHints.${value}`),
            tip: value === "audit" ? t("enforcementTips.audit") : undefined,
          }))}
        />
        {form.type === "semantic_policy" ? <SemanticFailureNote enforcement={form.enforcement} /> : null}
      </section>

      {form.type === "semantic_policy" ? <SemanticThresholds form={form} onChange={onChange} /> : null}

      <AppSwitchField
        id="guardrail-enabled"
        label={mode === "edit" ? t("form.enabled") : t("wizard.enabledLabel")}
        description={t("wizard.enabledHint")}
        checked={form.enabled}
        onCheckedChange={(enabled) => onChange({ enabled })}
      />
    </div>
  );
}
