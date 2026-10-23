import type { PolicyRule } from "./policy-rules";
import { readRuleTargets, subjectLabel, type RuleTargets } from "./rule-targets";

/** A translator bound to the `gatewayPolicies` namespace. */
export type SentenceTranslator = (key: string, values?: Record<string, string | number>) => string;

export type SentenceFormat = {
  t: SentenceTranslator;
  formatNumber: (value: number) => string;
};

function quoteList(values: string[], t: SentenceTranslator): string {
  return values.map((value) => t("sentence.quoted", { value })).join(t("sentence.listSeparator"));
}

function subjectPhrase(targets: RuleTargets, t: SentenceTranslator): string {
  const subject = targets.subject;
  if (!subject || subject.ids.length === 0) return t("sentence.anySubject");
  const names = subject.ids.map((id) => subjectLabel(subject, id));
  if (names.length === 1) {
    return t("sentence.subjectOne", { kind: t(`subjectKinds.${subject.kind}.one`), name: names[0] });
  }
  return t("sentence.subjectMany", { kind: t(`subjectKinds.${subject.kind}.many`), names: quoteList(names, t) });
}

function modelPhrase(targets: RuleTargets, t: SentenceTranslator): string {
  if (!targets.models || targets.models.length === 0) return t("sentence.anyModel");
  if (targets.models.length === 1) return targets.models[0];
  return t("sentence.modelMany", { models: targets.models.join(t("sentence.listSeparator")) });
}

function limitPhrase(item: PolicyRule, format: SentenceFormat): string {
  const { t, formatNumber } = format;
  if (item.type === "rate") {
    const { max, dimension, window } = item.rule.limit;
    return t("sentence.rateLimit", {
      max: formatNumber(max),
      dimension: t(`sentence.dimensions.${dimension}`),
      window: window.value === 1
        ? t(`sentence.units.${window.unit}`)
        : t("sentence.windowMany", { value: formatNumber(window.value), unit: t(`sentence.unitsMany.${window.unit}`) }),
    });
  }

  const period = t(`sentence.periods.${item.rule.unit}`);
  const scope = item.rule.budget_applies_per?.trim();
  const perScope = !scope || scope === "tenant"
    ? ""
    : t("sentence.appliesPer", {
        scope: scope.startsWith("metadata.") ? t("sentence.scopes.metadata", { key: scope.slice("metadata.".length) }) : t(`sentence.scopes.${scope}`),
      });
  return t("sentence.budgetLimit", { amount: formatNumber(Number(item.rule.limit ?? 0)), period }) + perScope;
}

/**
 * The one-sentence description of a rule, shared by the list, the editor
 * preview, and the delete confirmation, e.g. «برای تیم «چک‌اوت» روی
 * gpt-4o-mini حداکثر ۶۰۰ درخواست در هر دقیقه».
 */
export function ruleSentence(item: PolicyRule, format: SentenceFormat): string {
  const targets = readRuleTargets(item.rule.conditions?.conditions);
  const metadataCount = targets.metadata.length;
  return format.t("sentence.rule", {
    subjects: subjectPhrase(targets, format.t),
    models: modelPhrase(targets, format.t),
    limit: limitPhrase(item, format),
    metadata: metadataCount > 0 ? format.t("sentence.metadata", { count: metadataCount }) : "",
  });
}
