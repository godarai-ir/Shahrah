import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import QuickStartManagement from "./QuickStartManagement";
import type { QuickStartResources } from "./journey";

vi.mock("next-intl", () => ({
  useLocale: () => "en-US",
  useTranslations: (namespace: string) => (key: string, values?: Record<string, unknown>) => {
    if (values) return `${namespace}.${key}:${JSON.stringify(values)}`;
    return `${namespace}.${key}`;
  },
}));

vi.mock("@/features/quick-start/api", () => ({
  createPersonalAccessToken: vi.fn(),
  listQuickStartProviders: vi.fn(() => Promise.resolve({ data: [] })),
  runQuickStartSample: vi.fn(),
}));

vi.mock("@/lib/tenant-api-origin", () => ({
  getTenantApiBaseURL: () => "https://acme.godarai.test",
}));

vi.mock("@/components/app-ui/CodeBlock", () => ({
  CodeBlock: ({ code }: { code: string }) => <pre>{code}</pre>,
}));

vi.mock("@/components/app-ui/CodePanel", () => ({
  CodePanel: ({ code }: { code: string }) => <pre>{code}</pre>,
}));

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

function renderQuickStart(resources?: Partial<QuickStartResources>) {
  return render(
    <QuickStartManagement
      initialSubscription={{ plan_slug: "developer" }}
      initialEntitlements={{ features: { "auth.pat.enabled": true } }}
      initialPermissions={["provider_account:create", "settings:list"]}
      initialResources={makeResources({
        providerAccountCount: 1,
        providerAccountLabel: "openai:production",
        allowedPatModels: ["openai:default:gpt-4o-mini"],
        ...resources,
      })}
      docsBaseHref="https://godarai.test/docs"
    />,
  );
}

function stepItems(): HTMLElement[] {
  return Array.from(document.querySelectorAll("ol > li"));
}

describe("QuickStartManagement", () => {
  it("opens a single step and collapses the rest", () => {
    renderQuickStart();

    const items = stepItems();
    expect(items).toHaveLength(4);

    // Only the open step carries a description and a primary action.
    const withDescription = items.filter((item) =>
      within(item).queryByText(/^quickStart\.steps\..+\.description$/),
    );
    expect(withDescription).toHaveLength(1);
    expect(within(withDescription[0]).getByText("quickStart.steps.create_pat.title")).toBeTruthy();
    expect(within(withDescription[0]).getByText("quickStart.steps.skip")).toBeTruthy();
  });

  it("shows what completed a done step instead of only counting it", () => {
    renderQuickStart();

    const doneRow = stepItems()[0];
    expect(within(doneRow).getByText("openai:production")).toBeTruthy();
    expect(within(doneRow).getByText("quickStart.steps.donePrefix")).toBeTruthy();
    expect(within(doneRow).queryByText("quickStart.steps.connect_provider.description")).toBeNull();
  });

  it("explains why a collapsed step cannot be started", () => {
    const { unmount } = renderQuickStart();

    const items = stepItems();
    expect(within(items[2]).getByText('quickStart.locked.afterStep:{"index":2}')).toBeTruthy();
    unmount();

    // Without telemetry permission the last step is out of reach rather than queued.
    render(
      <QuickStartManagement
        initialSubscription={{ plan_slug: "developer" }}
        initialEntitlements={{ features: { "auth.pat.enabled": true } }}
        initialPermissions={["provider_account:create"]}
        initialResources={makeResources({
          providerAccountCount: 1,
          allowedPatModels: ["openai:default:gpt-4o-mini"],
        })}
        docsBaseHref="https://godarai.test/docs"
      />,
    );

    const blocked = stepItems()[3];
    expect(within(blocked).getByText("quickStart.locked.permission")).toBeTruthy();
    expect(screen.getByText('quickStart.progress.value:{"done":1,"total":3}')).toBeTruthy();
  });

  it("hands the open slot to the next step when one is skipped", () => {
    renderQuickStart();

    fireEvent.click(screen.getByText("quickStart.steps.skip"));

    const items = stepItems();
    expect(within(items[1]).getByText("quickStart.locked.skipped")).toBeTruthy();
    expect(within(items[2]).getByText("quickStart.steps.send_first_request.description")).toBeTruthy();
    // The skipped step stays reachable and stays in the denominator.
    expect(within(items[1]).getByText("quickStart.steps.resume")).toBeTruthy();
    expect(screen.getByText('quickStart.progress.value:{"done":1,"total":4}')).toBeTruthy();
  });

  it("renders the test bench in the page with one token type control", () => {
    renderQuickStart();

    const testBench = document.getElementById("test_bench");
    expect(testBench).not.toBeNull();
    expect(within(testBench as HTMLElement).getByText("quickStart.lab.tokenLocalOnly")).toBeTruthy();
    expect(within(testBench as HTMLElement).getAllByRole("tab", { name: "quickStart.tokenLabels.pat" })).toHaveLength(1);
  });

  it("caps the follow-up chips at four", () => {
    renderQuickStart();

    const chipBar = screen.getByText("quickStart.explore.afterJourney").parentElement as HTMLElement;
    expect(within(chipBar).getAllByRole("link").length).toBeLessThanOrEqual(4);
  });
});
