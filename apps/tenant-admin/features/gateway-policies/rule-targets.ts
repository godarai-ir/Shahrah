import type { RateLimitCondition } from "./rate-limiting/ruleTypes";

/**
 * Rate and budget conditions share one shape; the gateway evaluates both with
 * the same matcher (billing.budgetRuleMatches).
 */
export type PolicyCondition = RateLimitCondition;

/** The identity a subject condition compares against. The gateway defaults to user when unset. */
export type SubjectKind = NonNullable<PolicyCondition["subjectKind"]>;

export const SUBJECT_KINDS: readonly SubjectKind[] = ["user", "team", "virtual_account"];

export type SubjectPickerKind = "user" | "team" | "virtual_account";

export function subjectKindFromPicker(kind: string): SubjectKind | null {
  if (kind === "user" || kind === "team" || kind === "virtual_account") return kind;
  return null;
}

export function pickerKindFromSubject(kind: SubjectKind): SubjectPickerKind {
  return kind;
}

export type SubjectTarget = {
  kind: SubjectKind;
  ids: string[];
  labels: Record<string, string>;
};

export type RuleTargets = {
  /** null means every requester. */
  subject: SubjectTarget | null;
  /** null means every model. */
  models: string[] | null;
  metadata: PolicyCondition[];
  /**
   * Subject values were saved as `kind:id` tokens without a subjectKind. The
   * gateway compares the raw token with a user ID, so the rule never matches.
   */
  legacySubject: boolean;
  /** A legacy subject mixed several kinds; only the first kind survives a re-save. */
  droppedLegacyKinds: boolean;
  /**
   * Further subject or model conditions beyond the first of each (only API
   * clients create these). The editor does not show them, but they are written
   * back unchanged so a save never widens the rule.
   */
  extra?: PolicyCondition[];
};

const LEGACY_TOKEN = /^(user|team|virtual_account|external_identity|everyone):(.+)$/;

function cleanValues(values: string[] | undefined): string[] {
  return [...new Set((values ?? []).map((value) => value.trim()).filter(Boolean))];
}

function readLegacySubject(values: string[]): Pick<RuleTargets, "subject" | "droppedLegacyKinds"> | null {
  const tokens = values.map((value) => LEGACY_TOKEN.exec(value));
  if (tokens.length === 0 || tokens.some((token) => token === null)) return null;

  const parsed = tokens.map((token) => ({ kind: subjectKindFromPicker(token![1]), id: token![2] }));
  // Without a kind the gateway understands there is nothing to upgrade to; keeping
  // the raw tokens preserves "matches nothing" instead of dropping to every requester.
  const kind = parsed.find((item) => item.kind)?.kind;
  if (!kind) return null;

  const kept = parsed.filter((item) => item.kind === kind);
  return {
    subject: { kind, ids: kept.map((item) => item.id), labels: {} },
    droppedLegacyKinds: kept.length !== parsed.length,
  };
}

/** Reads the conditions the rule editor can express. */
export function readRuleTargets(conditions: PolicyCondition[] | undefined): RuleTargets {
  const targets: RuleTargets = { subject: null, models: null, metadata: [], legacySubject: false, droppedLegacyKinds: false, extra: [] };
  let seenSubject = false;
  let seenModel = false;

  for (const condition of conditions ?? []) {
    if (condition.field === "metadata") {
      targets.metadata.push(condition);
      continue;
    }

    if ((condition.field === "model" && seenModel) || (condition.field === "subject" && seenSubject)) {
      targets.extra!.push(condition);
      continue;
    }

    const values = cleanValues(condition.values);
    if (condition.field === "model") {
      seenModel = true;
      targets.models = values;
      continue;
    }

    seenSubject = true;

    const legacy = condition.subjectKind ? null : readLegacySubject(values);
    if (legacy) {
      targets.legacySubject = true;
      targets.droppedLegacyKinds = legacy.droppedLegacyKinds;
      targets.subject = legacy.subject;
      continue;
    }

    targets.subject = { kind: condition.subjectKind ?? "user", ids: values, labels: condition.valueLabels ?? {} };
  }

  return targets;
}

/** Writes targets back in the shape the gateway reads: raw IDs plus subjectKind, one condition per group. */
export function buildRuleConditions(targets: Pick<RuleTargets, "subject" | "models" | "metadata" | "extra">): PolicyCondition[] {
  const conditions: PolicyCondition[] = [];

  if (targets.subject && targets.subject.ids.length > 0) {
    const labels = Object.fromEntries(
      targets.subject.ids
        .filter((id) => targets.subject!.labels[id] && targets.subject!.labels[id] !== id)
        .map((id) => [id, targets.subject!.labels[id]]),
    );
    conditions.push({
      field: "subject",
      operator: "in",
      subjectKind: targets.subject.kind,
      values: targets.subject.ids,
      ...(Object.keys(labels).length > 0 ? { valueLabels: labels } : {}),
    });
  }

  if (targets.models && targets.models.length > 0) {
    conditions.push({ field: "model", operator: "in", values: targets.models });
  }

  return [...conditions, ...(targets.extra ?? []), ...targets.metadata];
}

export function subjectLabel(subject: SubjectTarget, id: string): string {
  return subject.labels[id] || id;
}
