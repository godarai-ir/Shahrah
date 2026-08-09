export type CommercialPlanSlug = "developer" | "team" | "growth" | "enterprise";

export type QuickStartSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status?: string;
  billing_interval?: string;
} | null;

export type QuickStartEntitlements = {
  features?: Record<string, unknown>;
} | null;

export type QuickStartResources = {
  providerAccountCount: number;
  personalAccessTokenCount: number;
  virtualUserCount: number;
  teamCount: number;
  orgUnitCount: number;
  rateLimitRuleCount: number;
  budgetLimitRuleCount: number;
  guardrailCount: number;
  allowedPatModels: string[];
  allowedVatModels: string[];
  sampleRequestSucceeded?: boolean;
  telemetryVisited?: boolean;
};

export type QuickStartStepID =
  | "connect_provider"
  | "create_pat"
  | "create_service_identity"
  | "create_rate_limit_rule"
  | "create_budget_limit_rule"
  | "create_guardrail"
  | "create_team"
  | "create_org_unit"
  | "send_first_request"
  | "inspect_activity";

export type QuickStartStepStatus = "done" | "active" | "locked" | "unavailable";

export type QuickStartAction =
  | { type: "link"; href: string }
  | { type: "dialog"; dialog: "create_pat" }
  | { type: "anchor"; anchor: "test_bench" };

export type QuickStartExploreCardID =
  | "provider_accounts"
  | "virtual_users"
  | "teams"
  | "org_units"
  | "guardrails"
  | "prompts"
  | "metrics"
  | "logs"
  | "billing_plan";

export type QuickStartJourneyStep = {
  id: QuickStartStepID;
  status: QuickStartStepStatus;
  action: QuickStartAction;
  blockedBy?: "feature" | "permission" | "prerequisite";
};

export type QuickStartJourney = {
  plan: CommercialPlanSlug;
  doneCount: number;
  totalCount: number;
  recommendedTokenType: "pat" | "vat";
  steps: QuickStartJourneyStep[];
  exploreCards: QuickStartExploreCardID[];
  testBench: {
    availableTokenTypes: Array<"pat" | "vat">;
    ready: boolean;
    recommendedTokenType: "pat" | "vat";
    suggestedModelByTokenType: {
      pat: string | null;
      vat: string | null;
    };
  };
};

type DeriveJourneyInput = {
  entitlements: QuickStartEntitlements;
  permissions: string[];
  resources: QuickStartResources;
  subscription: QuickStartSubscription;
};

const FEATURE_PAT_ENABLED = "auth.pat.enabled";
const FEATURE_VAT_ENABLED = "auth.vat.enabled";
const FEATURE_RATE_LIMIT_RULES_ENABLED = "rate_limit_rules.enabled";
const FEATURE_BUDGET_LIMIT_RULES_ENABLED = "budget_limit_rules.enabled";
const FEATURE_TEAMS_ENABLED = "teams.enabled";
const FEATURE_GUARDRAIL_WEBHOOK_ENABLED = "guardrails.webhook.enabled";
const FEATURE_ORG_UNITS_ENABLED = "org_units.enabled";

export function deriveQuickStartJourney(input: DeriveJourneyInput): QuickStartJourney {
  const plan = normalizePlanSlug(input.subscription?.plan_slug);
  const permissions = normalizePermissions(input.permissions);
  const features = input.entitlements?.features ?? {};
  const resources = input.resources;

  const canManageProvider = hasAnyPermission(permissions, [
    "provider_account:create",
    "provider_account:manage",
    "provider_account:read",
  ]);
  const canManagePolicyList = hasAnyPermission(permissions, ["policy:list", "policy:manage"]);
  const canManagePolicyCreate = hasAnyPermission(permissions, ["policy:manage"]);
  const canManageVirtualUsers = hasAnyPermission(permissions, ["virtual_account:create", "virtual_account:manage"]);
  const canManageTeams = hasAnyPermission(permissions, ["team:create", "team:manage", "team:read"]);
  const canCreateTeams = hasAnyPermission(permissions, ["team:create", "team:manage"]);
  const canCreateOrgUnits = hasAnyPermission(permissions, ["org_unit:create", "org_unit:manage"]);
  const canInspectTelemetry = hasAnyPermission(permissions, ["settings:list", "logs:read_metadata", "logs:read_content"]);
  const canManagePat = readBooleanFeature(features[FEATURE_PAT_ENABLED], true);
  const canUseVatFeature = readBooleanFeature(features[FEATURE_VAT_ENABLED], false);
  const canUseRateLimitFeature = readBooleanFeature(features[FEATURE_RATE_LIMIT_RULES_ENABLED], false);
  const canUseBudgetLimitFeature = readBooleanFeature(features[FEATURE_BUDGET_LIMIT_RULES_ENABLED], false);
  const canUseTeamsFeature = readBooleanFeature(features[FEATURE_TEAMS_ENABLED], false);
  const canUseOrgUnitsFeature = readBooleanFeature(features[FEATURE_ORG_UNITS_ENABLED], false);
  const canUseGuardrailWebhookFeature = readBooleanFeature(features[FEATURE_GUARDRAIL_WEBHOOK_ENABLED], false);

  const availableTokenTypes: Array<"pat" | "vat"> = [];
  if (canManagePat) availableTokenTypes.push("pat");
  if (canUseVatFeature) availableTokenTypes.push("vat");

  const recommendedTokenType =
    plan === "developer" ? "pat" : resources.virtualUserCount > 0 || canUseVatFeature ? "vat" : "pat";

  const suggestedPatModel = resources.allowedPatModels[0] ?? null;
  const suggestedVatModel = resources.allowedVatModels[0] ?? null;
  const testBenchReady = resources.providerAccountCount > 0 && Boolean(suggestedPatModel || suggestedVatModel);

  const inspectHref = canInspectTelemetry
    ? permissions.has("logs:read_metadata") || permissions.has("logs:read_content")
      ? "/logs"
      : "/traces"
    : "/metrics";

  const stepConfigs: Record<CommercialPlanSlug, Array<{ id: QuickStartStepID; requirement: () => boolean; available: () => boolean }>> = {
    developer: [
      {
        id: "connect_provider",
        requirement: () => resources.providerAccountCount > 0,
        available: () => canManageProvider,
      },
      {
        id: "create_pat",
        requirement: () => resources.personalAccessTokenCount > 0,
        available: () => canManagePat,
      },
      {
        id: "send_first_request",
        requirement: () => Boolean(resources.sampleRequestSucceeded),
        available: () => testBenchReady,
      },
      {
        id: "inspect_activity",
        requirement: () => Boolean(resources.telemetryVisited),
        available: () => canInspectTelemetry,
      },
    ],
    team: [
      {
        id: "connect_provider",
        requirement: () => resources.providerAccountCount > 0,
        available: () => canManageProvider,
      },
      {
        id: "create_service_identity",
        requirement: () => resources.virtualUserCount > 0,
        available: () => canUseVatFeature && canManageVirtualUsers,
      },
      {
        id: "create_rate_limit_rule",
        requirement: () => resources.rateLimitRuleCount > 0,
        available: () => canUseRateLimitFeature && canManagePolicyCreate,
      },
      {
        id: "create_budget_limit_rule",
        requirement: () => resources.budgetLimitRuleCount > 0,
        available: () => canUseBudgetLimitFeature && canManagePolicyCreate,
      },
      {
        id: "send_first_request",
        requirement: () => Boolean(resources.sampleRequestSucceeded),
        available: () => testBenchReady,
      },
      {
        id: "inspect_activity",
        requirement: () => Boolean(resources.telemetryVisited),
        available: () => canInspectTelemetry,
      },
    ],
    growth: [
      {
        id: "connect_provider",
        requirement: () => resources.providerAccountCount > 0,
        available: () => canManageProvider,
      },
      {
        id: "create_service_identity",
        requirement: () => resources.virtualUserCount > 0,
        available: () => canUseVatFeature && canManageVirtualUsers,
      },
      {
        id: "create_rate_limit_rule",
        requirement: () => resources.rateLimitRuleCount > 0,
        available: () => canUseRateLimitFeature && canManagePolicyCreate,
      },
      {
        id: "create_budget_limit_rule",
        requirement: () => resources.budgetLimitRuleCount > 0,
        available: () => canUseBudgetLimitFeature && canManagePolicyCreate,
      },
      {
        id: "create_guardrail",
        requirement: () => resources.guardrailCount > 0,
        available: () => canUseGuardrailWebhookFeature && canManagePolicyList,
      },
      {
        id: "send_first_request",
        requirement: () => Boolean(resources.sampleRequestSucceeded),
        available: () => testBenchReady,
      },
      {
        id: "inspect_activity",
        requirement: () => Boolean(resources.telemetryVisited),
        available: () => canInspectTelemetry,
      },
    ],
    enterprise: [
      {
        id: "connect_provider",
        requirement: () => resources.providerAccountCount > 0,
        available: () => canManageProvider,
      },
      {
        id: "create_org_unit",
        requirement: () => resources.orgUnitCount > 0,
        available: () => canUseOrgUnitsFeature && canCreateOrgUnits,
      },
      {
        id: "create_team",
        requirement: () => resources.teamCount > 0,
        available: () => canUseTeamsFeature && canCreateTeams,
      },
      {
        id: "create_service_identity",
        requirement: () => resources.virtualUserCount > 0,
        available: () => canUseVatFeature && canManageVirtualUsers,
      },
      {
        id: "create_guardrail",
        requirement: () => resources.guardrailCount > 0,
        available: () => canUseGuardrailWebhookFeature && canManagePolicyList,
      },
      {
        id: "send_first_request",
        requirement: () => Boolean(resources.sampleRequestSucceeded),
        available: () => testBenchReady,
      },
      {
        id: "inspect_activity",
        requirement: () => Boolean(resources.telemetryVisited),
        available: () => canInspectTelemetry,
      },
    ],
  };

  let allPreviousDone = true;
  const steps = stepConfigs[plan].map((config) => {
    const done = config.requirement();
    const available = config.available();
    const status = done
      ? "done"
      : !available
        ? "unavailable"
        : allPreviousDone
          ? "active"
          : "locked";
    const step = {
      id: config.id,
      status,
      action: actionForStep(config.id, resources, inspectHref),
      blockedBy:
        status === "unavailable"
          ? isFeatureBlocked(config.id, {
              canManagePat,
              canUseVatFeature,
              canUseRateLimitFeature,
              canUseBudgetLimitFeature,
              canUseGuardrailWebhookFeature,
              canUseOrgUnitsFeature,
              canUseTeamsFeature,
            })
            ? "feature"
            : "permission"
          : status === "locked"
            ? "prerequisite"
            : undefined,
    } satisfies QuickStartJourneyStep;
    allPreviousDone = allPreviousDone && done;
    return step;
  });

  const exploreCards = buildExploreCards({
    canInspectTelemetry,
    canManagePolicyList,
    canManageProvider,
    canManageTeams,
    canManageVirtualUsers,
    canUseGuardrailWebhookFeature,
    canUseOrgUnitsFeature,
    canUseTeamsFeature,
  });

  return {
    plan,
    doneCount: steps.filter((step) => step.status === "done").length,
    totalCount: steps.length,
    recommendedTokenType,
    steps,
    exploreCards,
    testBench: {
      availableTokenTypes,
      ready: testBenchReady,
      recommendedTokenType,
      suggestedModelByTokenType: {
        pat: suggestedPatModel,
        vat: suggestedVatModel,
      },
    },
  };
}

function normalizePlanSlug(value?: string): CommercialPlanSlug {
  switch ((value ?? "").trim().toLowerCase()) {
    case "team":
      return "team";
    case "growth":
      return "growth";
    case "enterprise":
      return "enterprise";
    case "developer":
    default:
      return "developer";
  }
}

function normalizePermissions(permissions: string[]): Set<string> {
  return new Set(permissions.map((item) => item.trim().toLowerCase()).filter(Boolean));
}

function hasAnyPermission(permissions: Set<string>, required: string[]): boolean {
  return required.some((item) => permissions.has(item.trim().toLowerCase()));
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

function actionForStep(
  stepID: QuickStartStepID,
  resources: QuickStartResources,
  inspectHref: string,
): QuickStartAction {
  switch (stepID) {
    case "connect_provider":
      return {
        type: "link",
        href: resources.providerAccountCount > 0 ? "/provider-accounts" : "/provider-accounts/new",
      };
    case "create_pat":
      return {
        type: "dialog",
        dialog: "create_pat",
      };
    case "create_service_identity":
      return {
        type: "link",
        href: "/virtual-users",
      };
    case "create_rate_limit_rule":
      return {
        type: "link",
        href: resources.rateLimitRuleCount > 0
          ? "/gateway-policies/rate-limiting-rules"
          : "/gateway-policies/rate-limiting-rules/new",
      };
    case "create_budget_limit_rule":
      return {
        type: "link",
        href: resources.budgetLimitRuleCount > 0
          ? "/gateway-policies/budget-limiting-rules"
          : "/gateway-policies/budget-limiting-rules/new",
      };
    case "create_guardrail":
      return {
        type: "link",
        href: "/guardrails",
      };
    case "create_team":
      return {
        type: "link",
        href: "/teams",
      };
    case "create_org_unit":
      return {
        type: "link",
        href: "/org-units",
      };
    case "send_first_request":
      return {
        type: "anchor",
        anchor: "test_bench",
      };
    case "inspect_activity":
      return {
        type: "link",
        href: inspectHref,
      };
  }
}

function isFeatureBlocked(
  stepID: QuickStartStepID,
  flags: {
    canManagePat: boolean;
    canUseVatFeature: boolean;
    canUseRateLimitFeature: boolean;
    canUseBudgetLimitFeature: boolean;
    canUseGuardrailWebhookFeature: boolean;
    canUseOrgUnitsFeature: boolean;
    canUseTeamsFeature: boolean;
  },
): boolean {
  switch (stepID) {
    case "create_pat":
      return !flags.canManagePat;
    case "create_service_identity":
      return !flags.canUseVatFeature;
    case "create_rate_limit_rule":
      return !flags.canUseRateLimitFeature;
    case "create_budget_limit_rule":
      return !flags.canUseBudgetLimitFeature;
    case "create_guardrail":
      return !flags.canUseGuardrailWebhookFeature;
    case "create_org_unit":
      return !flags.canUseOrgUnitsFeature;
    case "create_team":
      return !flags.canUseTeamsFeature;
    default:
      return false;
  }
}

function buildExploreCards(flags: {
  canInspectTelemetry: boolean;
  canManagePolicyList: boolean;
  canManageProvider: boolean;
  canManageTeams: boolean;
  canManageVirtualUsers: boolean;
  canUseGuardrailWebhookFeature: boolean;
  canUseOrgUnitsFeature: boolean;
  canUseTeamsFeature: boolean;
}): QuickStartExploreCardID[] {
  const cards: QuickStartExploreCardID[] = ["billing_plan"];
  if (flags.canManageProvider) cards.push("provider_accounts");
  if (flags.canManageVirtualUsers) cards.push("virtual_users");
  if (flags.canUseTeamsFeature && flags.canManageTeams) cards.push("teams");
  if (flags.canUseOrgUnitsFeature) cards.push("org_units");
  if (flags.canManagePolicyList) cards.push("prompts");
  if (flags.canUseGuardrailWebhookFeature && flags.canManagePolicyList) cards.push("guardrails");
  if (flags.canInspectTelemetry) {
    cards.push("metrics");
    cards.push("logs");
  }
  return Array.from(new Set(cards));
}
