"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { ChipInput } from "@/components/app-ui/ChipInput";
import { ModelEntitySelect, type ModelSelection } from "@/components/app-ui/ModelEntitySelect";
import { PrincipalEntitySelect } from "@/components/app-ui/PrincipalEntitySelect";
import { RemovableChipList } from "@/components/app-ui/RemovableChipList";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Badge } from "@/components/ui/badge";
import { FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Separator } from "@/components/ui/separator";

import { loadPolicyModelOptions, searchPolicyPrincipals } from "./lookups";
import { policyRuleKey, policyRuleName, type PolicyRule } from "./policy-rules";
import type { GatewayRequestContext } from "./rate-limiting/ruleTypes";
import { SUBJECT_KINDS, subjectKindFromPicker, type SubjectKind } from "./rule-targets";
import { simulateRequest, type BudgetOutcome, type RateOutcome } from "./simulate-request";

type Identity = { id: string; label: string };

const IDENTITY_FIELD: Record<SubjectKind, keyof GatewayRequestContext> = {
  user: "userId",
  team: "teamId",
  virtual_account: "apiKeyId",
};

const RATE_OUTCOME_STATUS: Record<RateOutcome, "success" | "warning" | "neutral"> = {
  enforced: "success",
  countOnly: "warning",
  notReached: "neutral",
};

const BUDGET_OUTCOME_STATUS: Record<BudgetOutcome, "success" | "warning"> = {
  enforced: "success",
  alertOnly: "warning",
};

function parseMetadataEntries(raw: string): string[] {
  return raw.split(/[,\n]+/).map((item) => item.trim()).filter((item) => /^[^=\s]+=.*$/.test(item));
}

type RequestSimulatorProps = {
  rules: PolicyRule[];
  onOpenRule: (item: PolicyRule) => void;
};

/** Replays the gateway's rule pass for a hand-built request, entirely in the browser. */
export function RequestSimulator({ rules, onOpenRule }: RequestSimulatorProps) {
  const t = useTranslations("gatewayPolicies");
  const [identities, setIdentities] = useState<Partial<Record<SubjectKind, Identity>>>({});
  const [model, setModel] = useState<ModelSelection | null>(null);
  const [metadata, setMetadata] = useState<string[]>([]);
  const [models, setModels] = useState<{ providerAllowed: ModelSelection[]; virtual: ModelSelection[] }>({ providerAllowed: [], virtual: [] });
  const [now] = useState(() => new Date());

  useEffect(() => {
    let active = true;
    void loadPolicyModelOptions().then((options) => {
      if (active) setModels({ providerAllowed: options.providerAllowedModels, virtual: options.virtualModels });
    });
    return () => {
      active = false;
    };
  }, []);

  const request = useMemo<GatewayRequestContext>(() => {
    const context: GatewayRequestContext = { model: model?.name, metadata: {} };
    for (const kind of SUBJECT_KINDS) {
      const identity = identities[kind];
      if (identity) Object.assign(context, { [IDENTITY_FIELD[kind]]: identity.id });
    }
    for (const entry of metadata) {
      const separator = entry.indexOf("=");
      context.metadata![entry.slice(0, separator)] = entry.slice(separator + 1);
    }
    return context;
  }, [identities, metadata, model]);

  const result = useMemo(() => simulateRequest(rules, request, now), [now, request, rules]);
  const byRuleID = useMemo(() => new Map(rules.map((item) => [policyRuleKey(item), item])), [rules]);
  const open = (type: PolicyRule["type"], ruleID: string) => {
    const item = byRuleID.get(`${type}:${ruleID}`);
    if (item) onOpenRule(item);
  };
  const nothingApplies = result.rate.length === 0 && result.budget.length === 0;
  const identityChips = SUBJECT_KINDS.flatMap((kind) => {
    const identity = identities[kind];
    if (!identity) return [];
    const label = t("editor.subjectChip", { kind: t(`subjectKinds.${kind}.one`), name: identity.label });
    return [{ value: kind, label, removeLabel: t("editor.remove", { label }) }];
  });

  return (
    <AppCard title={t("simulator.title")} description={t("simulator.description")} contentClassName="flex flex-col gap-4">
      <FieldGroup className="gap-4">
        <div className="flex flex-col gap-2">
          <FieldLabel htmlFor="simulator-subject">{t("simulator.requester")}</FieldLabel>
          <PrincipalEntitySelect
            id="simulator-subject"
            selected={null}
            kinds={["user", "team", "virtual_account"]}
            onSearch={searchPolicyPrincipals}
            triggerClassName="h-9 min-h-9 rounded-lg"
            labels={{
              users: t("editor.picker.users"),
              teams: t("editor.picker.teams"),
              virtualAccounts: t("editor.picker.virtualAccounts"),
              externalIdentities: t("editor.picker.externalIdentities"),
              everyone: t("targets.anySubject"),
              searchPlaceholder: t("editor.picker.searchSubject"),
              triggerPlaceholder: t("simulator.pickRequester"),
              loading: t("editor.picker.loading"),
              empty: t("editor.picker.empty"),
              minSearchHint: t("editor.picker.minSearchHint"),
            }}
            onSelectedChange={(selection) => {
              const kind = selection ? subjectKindFromPicker(selection.type) : null;
              const id = selection?.principal_id?.trim();
              if (!selection || !kind || !id) return;
              setIdentities((current) => ({ ...current, [kind]: { id, label: selection.display_name || id } }));
            }}
          />
          {identityChips.length > 0 ? (
            <RemovableChipList
              variant="outline"
              items={identityChips}
              onRemove={(kind) => setIdentities((current) => ({ ...current, [kind]: undefined }))}
            />
          ) : null}
          <FieldDescription>{t("simulator.requesterHint")}</FieldDescription>
        </div>

        <div className="flex flex-col gap-2">
          <FieldLabel htmlFor="simulator-model">{t("simulator.model")}</FieldLabel>
          <ModelEntitySelect
            id="simulator-model"
            selected={model}
            onSelectedChange={setModel}
            providerAllowedModels={models.providerAllowed}
            virtualModels={models.virtual}
            triggerClassName="h-9 rounded-lg"
            labels={{
              triggerPlaceholder: t("targets.anyModel"),
              searchPlaceholder: t("editor.picker.searchModel"),
              providerAllowed: t("editor.picker.providerModels"),
              virtualModels: t("editor.picker.virtualModels"),
              loading: t("editor.picker.loading"),
              empty: t("editor.picker.empty"),
            }}
          />
        </div>

        <div className="flex flex-col gap-2">
          <FieldLabel htmlFor="simulator-metadata">{t("simulator.metadata")}</FieldLabel>
          <ChipInput
            id="simulator-metadata"
            value={metadata}
            onValueChange={setMetadata}
            parseValues={parseMetadataEntries}
            placeholder={t("simulator.metadataPlaceholder")}
            getRemoveLabel={(value) => t("editor.remove", { label: value })}
          />
        </div>
      </FieldGroup>

      <Separator />

      <section aria-live="polite" aria-label={t("simulator.resultLabel")} className="flex flex-col gap-3">
        {nothingApplies ? <p className="text-sm leading-6 text-muted-foreground">{t("simulator.empty")}</p> : null}

        {result.rate.length > 0 ? (
          <ul className="flex flex-col gap-2">
            {result.rate.map(({ rule, outcome }) => (
              <li key={rule.rule_id} className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                  <Badge variant="secondary" className="font-mono tabular-nums" dir="ltr">{rule.priority}</Badge>
                  <button type="button" className="min-w-0 flex-1 truncate text-start text-sm font-medium hover:underline" onClick={() => open("rate", rule.rule_id)}>
                    {policyRuleName({ type: "rate", rule })}
                  </button>
                  <StatusBadge status={RATE_OUTCOME_STATUS[outcome]}>{t(`simulator.outcomes.${outcome}`)}</StatusBadge>
                </div>
                {outcome === "notReached" && result.stoppedBy ? (
                  <p className="text-xs leading-5 text-muted-foreground">
                    {t("simulator.notReachedNote", { name: policyRuleName({ type: "rate", rule: result.stoppedBy }) })}
                  </p>
                ) : null}
              </li>
            ))}
          </ul>
        ) : null}

        {result.budget.length > 0 ? (
          <ul className="flex flex-col gap-2">
            {result.budget.map(({ rule, outcome }) => (
              <li key={rule.rule_id} className="flex items-center gap-2">
                <Badge variant="outline">{t("types.budgetShort")}</Badge>
                <button type="button" className="min-w-0 flex-1 truncate text-start text-sm font-medium hover:underline" onClick={() => open("budget", rule.rule_id)}>
                  {policyRuleName({ type: "budget", rule })}
                </button>
                <StatusBadge status={BUDGET_OUTCOME_STATUS[outcome]}>{t(`simulator.outcomes.${outcome}`)}</StatusBadge>
              </li>
            ))}
          </ul>
        ) : null}

        {!nothingApplies ? <p className="text-xs leading-5 text-muted-foreground">{t("simulator.enforcementNote")}</p> : null}

        {result.dormant.length > 0 ? (
          <div className="flex flex-col gap-2 rounded-lg border border-dashed border-border p-3">
            <p className="text-xs font-medium text-muted-foreground">{t("simulator.dormantTitle")}</p>
            <ul className="flex flex-col gap-1">
              {result.dormant.map(({ item, reason }) => (
                <li key={policyRuleKey(item)} className="flex flex-col">
                  <button type="button" className="truncate text-start text-sm hover:underline" onClick={() => onOpenRule(item)}>
                    {policyRuleName(item)}
                  </button>
                  <span className="text-xs leading-5 text-muted-foreground">{t(`simulator.dormant.${reason}`)}</span>
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </section>
    </AppCard>
  );
}
