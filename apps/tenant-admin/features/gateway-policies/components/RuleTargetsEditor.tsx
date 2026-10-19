"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Plus, Trash2 } from "lucide-react";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { ModelEntitySelect, type ModelEntitySelectLabels, type ModelSelection } from "@/components/app-ui/ModelEntitySelect";
import { PrincipalEntitySelect, type PrincipalEntitySelectLabels } from "@/components/app-ui/PrincipalEntitySelect";
import { RemovableChipList } from "@/components/app-ui/RemovableChipList";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FieldDescription, FieldGroup, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { loadPolicyModelOptions, searchPolicyPrincipals } from "../lookups";
import {
  buildRuleConditions,
  pickerKindFromSubject,
  readRuleTargets,
  subjectKindFromPicker,
  subjectLabel,
  type PolicyCondition,
  type RuleTargets,
  type SubjectPickerKind,
} from "../rule-targets";

const ALL_PICKER_KINDS: SubjectPickerKind[] = ["team", "user", "virtual_account"];
const METADATA_OPERATORS = ["equals", "not_equals", "contains", "starts_with", "ends_with", "exists", "not_exists", "gt", "gte", "lt", "lte"] as const;
const VALUELESS_OPERATORS = new Set<string>(["exists", "not_exists"]);

type RuleTargetsEditorProps = {
  value: PolicyCondition[] | undefined;
  onChange: (conditions: PolicyCondition[]) => void;
  disabled?: boolean;
  /**
   * How the stored rule looked before the form normalized it. The form upgrades
   * legacy subjects on load, so `value` alone no longer shows that anything changed.
   */
  storedTargets?: Pick<RuleTargets, "legacySubject" | "droppedLegacyKinds">;
};

/**
 * Edits exactly what the gateway can evaluate: one set of requesters of a
 * single kind, one set of models, and metadata conditions, all joined by AND.
 */
export function RuleTargetsEditor({ value, onChange, disabled, storedTargets }: RuleTargetsEditorProps) {
  const t = useTranslations("gatewayPolicies");
  const targets = readRuleTargets(value);
  const [models, setModels] = useState<{ providerAllowed: ModelSelection[]; virtual: ModelSelection[] }>({ providerAllowed: [], virtual: [] });

  useEffect(() => {
    let active = true;
    void loadPolicyModelOptions().then((options) => {
      if (active) setModels({ providerAllowed: options.providerAllowedModels, virtual: options.virtualModels });
    });
    return () => {
      active = false;
    };
  }, []);

  function update(next: Partial<Pick<RuleTargets, "subject" | "models" | "metadata">>) {
    onChange(buildRuleConditions({ ...targets, ...next }));
  }

  const subject = targets.subject;
  const subjectPickerKinds = subject && subject.ids.length > 0 ? [pickerKindFromSubject(subject.kind)] : ALL_PICKER_KINDS;
  const principalLabels: PrincipalEntitySelectLabels = {
    users: t("editor.picker.users"),
    teams: t("editor.picker.teams"),
    virtualAccounts: t("editor.picker.virtualAccounts"),
    externalIdentities: t("editor.picker.externalIdentities"),
    everyone: t("targets.anySubject"),
    searchPlaceholder: t("editor.picker.searchSubject"),
    triggerPlaceholder: t("editor.picker.addSubject"),
    loading: t("editor.picker.loading"),
    empty: t("editor.picker.empty"),
    minSearchHint: t("editor.picker.minSearchHint"),
  };
  const modelLabels: ModelEntitySelectLabels = {
    triggerPlaceholder: t("editor.picker.addModel"),
    searchPlaceholder: t("editor.picker.searchModel"),
    providerAllowed: t("editor.picker.providerModels"),
    virtualModels: t("editor.picker.virtualModels"),
    loading: t("editor.picker.loading"),
    empty: t("editor.picker.empty"),
  };

  return (
    <FieldGroup className="gap-6">
      {storedTargets?.legacySubject || targets.legacySubject ? (
        <Alert>
          <AlertDescription>
            {t("editor.legacySubject")}
            {storedTargets?.droppedLegacyKinds || targets.droppedLegacyKinds ? ` ${t("editor.legacySubjectDropped")}` : null}
          </AlertDescription>
        </Alert>
      ) : null}

      <FieldSet>
        <FieldLegend variant="label">{t("editor.subjectTitle")}</FieldLegend>
        <FieldDescription>{t("editor.subjectDescription")}</FieldDescription>
        <div className="flex flex-col gap-3">
          {subject && subject.ids.length > 0 ? (
            <RemovableChipList
              variant="outline"
              disabled={disabled}
              items={subject.ids.map((id) => {
                const label = t("editor.subjectChip", { kind: t(`subjectKinds.${subject.kind}.one`), name: subjectLabel(subject, id) });
                return { value: id, label, removeLabel: t("editor.remove", { label }) };
              })}
              onRemove={(id) => update({ subject: { ...subject, ids: subject.ids.filter((item) => item !== id) } })}
            />
          ) : (
            <Badge variant="secondary" className="w-fit font-normal">{t("targets.anySubject")}</Badge>
          )}
          <PrincipalEntitySelect
            id="rule-subject-picker"
            selected={null}
            disabled={disabled}
            kinds={subjectPickerKinds}
            labels={principalLabels}
            onSearch={searchPolicyPrincipals}
            triggerClassName="h-10 min-h-10 rounded-lg"
            onSelectedChange={(selection) => {
              const kind = selection ? subjectKindFromPicker(selection.type) : null;
              const id = selection?.principal_id?.trim();
              if (!selection || !kind || !id) return;
              const current = subject && subject.kind === kind ? subject : { kind, ids: [], labels: {} };
              if (current.ids.includes(id)) return;
              update({
                subject: {
                  kind,
                  ids: [...current.ids, id],
                  labels: { ...current.labels, [id]: selection.display_name || id },
                },
              });
            }}
          />
        </div>
      </FieldSet>

      <FieldSet>
        <FieldLegend variant="label">{t("editor.modelTitle")}</FieldLegend>
        <FieldDescription>{t("editor.modelDescription")}</FieldDescription>
        <div className="flex flex-col gap-3">
          {targets.models && targets.models.length > 0 ? (
            <RemovableChipList
              variant="outline"
              disabled={disabled}
              items={targets.models.map((model) => ({ value: model, label: model, textDir: "ltr" as const, removeLabel: t("editor.remove", { label: model }) }))}
              onRemove={(model) => update({ models: (targets.models ?? []).filter((item) => item !== model) })}
            />
          ) : (
            <Badge variant="secondary" className="w-fit font-normal">{t("targets.anyModel")}</Badge>
          )}
          <ModelEntitySelect
            id="rule-model-picker"
            selected={null}
            providerAllowedModels={models.providerAllowed}
            virtualModels={models.virtual}
            labels={modelLabels}
            triggerClassName="h-10 rounded-lg"
            onSelectedChange={(selection) => {
              if (!selection?.name || targets.models?.includes(selection.name)) return;
              update({ models: [...(targets.models ?? []), selection.name] });
            }}
          />
        </div>
      </FieldSet>

      <FieldSet>
        <div className="flex items-center justify-between gap-3">
          <FieldLegend variant="label" className="mb-0">{t("editor.metadataTitle")}</FieldLegend>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={disabled}
            onClick={() => update({ metadata: [...targets.metadata, { field: "metadata", operator: "equals", metadataKey: "", value: "" }] })}
          >
            <Plus data-icon="inline-start" />
            {t("editor.addMetadata")}
          </Button>
        </div>
        {targets.metadata.length === 0 ? (
          <FieldDescription>{t("metadata.none")}</FieldDescription>
        ) : (
          <div className="flex flex-col gap-2">
            <FieldDescription>{t("editor.metadataDescription")}</FieldDescription>
            {targets.metadata.map((condition, index) => {
              const setRow = (patch: Partial<PolicyCondition>) =>
                update({ metadata: targets.metadata.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)) });
              const needsValue = !VALUELESS_OPERATORS.has(condition.operator);
              const keyMissing = !condition.metadataKey?.trim();
              const valueMissing = needsValue && `${condition.value ?? ""}`.trim() === "";
              return (
                <div key={index} className="grid gap-2 rounded-lg border border-border p-3 sm:grid-cols-[minmax(0,1fr)_10rem_minmax(0,1fr)_auto]">
                  <Input
                    dir="ltr"
                    className="font-mono"
                    aria-label={t("editor.metadataKey", { index: index + 1 })}
                    aria-invalid={keyMissing || undefined}
                    placeholder={t("editor.metadataKeyPlaceholder")}
                    value={condition.metadataKey ?? ""}
                    disabled={disabled}
                    onChange={(event) => setRow({ metadataKey: event.target.value })}
                  />
                  <AppSelectField
                    ariaLabel={t("editor.metadataOperator", { index: index + 1 })}
                    value={condition.operator}
                    disabled={disabled}
                    triggerClassName="min-w-0 rounded-lg"
                    onValueChange={(operator) => setRow({ operator: operator as PolicyCondition["operator"], ...(VALUELESS_OPERATORS.has(operator) ? { value: undefined } : {}) })}
                    options={METADATA_OPERATORS.map((operator) => ({ value: operator, label: t(`operators.${operator}`) }))}
                  />
                  {needsValue ? (
                    <Input
                      dir="ltr"
                      className="font-mono"
                      aria-label={t("editor.metadataValue", { index: index + 1 })}
                      aria-invalid={valueMissing || undefined}
                      placeholder={t("editor.metadataValuePlaceholder")}
                      value={`${condition.value ?? ""}`}
                      disabled={disabled}
                      onChange={(event) => setRow({ value: event.target.value })}
                    />
                  ) : (
                    <span aria-hidden="true" />
                  )}
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    disabled={disabled}
                    aria-label={t("editor.removeMetadata", { index: index + 1 })}
                    onClick={() => update({ metadata: targets.metadata.filter((_, rowIndex) => rowIndex !== index) })}
                  >
                    <Trash2 />
                  </Button>
                </div>
              );
            })}
          </div>
        )}
      </FieldSet>
    </FieldGroup>
  );
}
