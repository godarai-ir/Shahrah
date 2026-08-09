import { describe, expect, it } from "vitest";

import { deriveQuickStartJourney, type QuickStartResources } from "./journey";

function makeResources(partial?: Partial<QuickStartResources>): QuickStartResources {
  return {
    providerAccountCount: 0,
    personalAccessTokenCount: 0,
    virtualUserCount: 0,
    teamCount: 0,
    orgUnitCount: 0,
    rateLimitRuleCount: 0,
    budgetLimitRuleCount: 0,
    guardrailCount: 0,
    allowedPatModels: [],
    allowedVatModels: [],
    ...partial,
  };
}

describe("deriveQuickStartJourney", () => {
  it("activates PAT creation after provider setup for developer plan", () => {
    const journey = deriveQuickStartJourney({
      subscription: { plan_slug: "developer" },
      entitlements: { features: { "auth.pat.enabled": true } },
      permissions: ["provider_account:create"],
      resources: makeResources({
        providerAccountCount: 1,
        allowedPatModels: ["openai:default:gpt-4o-mini"],
      }),
    });

    expect(journey.plan).toBe("developer");
    expect(journey.steps.map((step) => [step.id, step.status])).toEqual([
      ["connect_provider", "done"],
      ["create_pat", "active"],
      ["send_first_request", "locked"],
      ["inspect_activity", "unavailable"],
    ]);
    expect(journey.recommendedTokenType).toBe("pat");
    expect(journey.testBench.ready).toBe(true);
  });

  it("keeps the team plan focused on service identity and policy setup", () => {
    const journey = deriveQuickStartJourney({
      subscription: { plan_slug: "team" },
      entitlements: {
        features: {
          "auth.pat.enabled": true,
          "auth.vat.enabled": true,
          "rate_limit_rules.enabled": true,
          "budget_limit_rules.enabled": true,
        },
      },
      permissions: ["provider_account:create", "virtual_account:create", "policy:manage", "settings:list"],
      resources: makeResources({
        providerAccountCount: 1,
        virtualUserCount: 1,
        rateLimitRuleCount: 1,
        allowedPatModels: ["openai:default:gpt-4o-mini"],
        allowedVatModels: ["openai:default:gpt-4o-mini"],
      }),
    });

    expect(journey.steps.map((step) => [step.id, step.status])).toEqual([
      ["connect_provider", "done"],
      ["create_service_identity", "done"],
      ["create_rate_limit_rule", "done"],
      ["create_budget_limit_rule", "active"],
      ["send_first_request", "locked"],
      ["inspect_activity", "locked"],
    ]);
    expect(journey.recommendedTokenType).toBe("vat");
    expect(journey.testBench.availableTokenTypes).toEqual(["pat", "vat"]);
  });

  it("adds a guardrail step for growth plan and marks it active after policies", () => {
    const journey = deriveQuickStartJourney({
      subscription: { plan_slug: "growth" },
      entitlements: {
        features: {
          "auth.pat.enabled": true,
          "auth.vat.enabled": true,
          "rate_limit_rules.enabled": true,
          "budget_limit_rules.enabled": true,
          "guardrails.webhook.enabled": true,
        },
      },
      permissions: ["provider_account:create", "virtual_account:create", "policy:manage", "policy:list", "settings:list"],
      resources: makeResources({
        providerAccountCount: 1,
        virtualUserCount: 1,
        rateLimitRuleCount: 1,
        budgetLimitRuleCount: 1,
        allowedPatModels: ["openai:default:gpt-4o-mini"],
        allowedVatModels: ["openai:default:gpt-4o-mini"],
      }),
    });

    expect(journey.steps.map((step) => [step.id, step.status])).toEqual([
      ["connect_provider", "done"],
      ["create_service_identity", "done"],
      ["create_rate_limit_rule", "done"],
      ["create_budget_limit_rule", "done"],
      ["create_guardrail", "active"],
      ["send_first_request", "locked"],
      ["inspect_activity", "locked"],
    ]);
    expect(journey.exploreCards).toContain("guardrails");
  });

  it("marks enterprise org-unit setup unavailable when permission is missing", () => {
    const journey = deriveQuickStartJourney({
      subscription: { plan_slug: "enterprise" },
      entitlements: {
        features: {
          "auth.pat.enabled": true,
          "auth.vat.enabled": true,
          "teams.enabled": true,
          "org_units.enabled": true,
          "guardrails.webhook.enabled": true,
        },
      },
      permissions: ["provider_account:create", "team:manage", "virtual_account:create", "policy:list"],
      resources: makeResources({
        providerAccountCount: 1,
      }),
    });

    expect(journey.steps[0]).toMatchObject({ id: "connect_provider", status: "done" });
    expect(journey.steps[1]).toMatchObject({
      id: "create_org_unit",
      status: "unavailable",
      blockedBy: "permission",
    });
    expect(journey.steps[2]).toMatchObject({ id: "create_team", status: "locked" });
  });

  it("marks sample request and telemetry steps done from local quick-start progress", () => {
    const journey = deriveQuickStartJourney({
      subscription: { plan_slug: "developer" },
      entitlements: { features: { "auth.pat.enabled": true } },
      permissions: ["provider_account:create", "logs:read_metadata"],
      resources: makeResources({
        providerAccountCount: 1,
        personalAccessTokenCount: 1,
        allowedPatModels: ["openai:default:gpt-4o-mini"],
        sampleRequestSucceeded: true,
        telemetryVisited: true,
      }),
    });

    expect(journey.steps.map((step) => [step.id, step.status])).toEqual([
      ["connect_provider", "done"],
      ["create_pat", "done"],
      ["send_first_request", "done"],
      ["inspect_activity", "done"],
    ]);
  });
});
