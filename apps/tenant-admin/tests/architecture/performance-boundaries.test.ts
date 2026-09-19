import { readFile } from "node:fs/promises";
import path from "node:path";

import { describe, expect, it } from "vitest";

const appRoot = process.cwd();

const componentBudgets = [
  ["features/provider-accounts/ProviderAccountWizard.tsx", 2_200],
  ["features/provider-accounts/ProviderAccountCreateView.tsx", 625],
  ["features/logs/LogsCommandCenter.tsx", 1_250],
  ["features/provider-accounts/models-editor.tsx", 1_250],
  ["features/traces/TraceManagement.tsx", 450],
  ["features/billing/BillingFinancePanels.tsx", 875],
  ["features/metrics/ModelMetricsDashboard.tsx", 600],
  ["features/dashboard/DashboardAnalytics.tsx", 550],
  ["features/virtual-models/VirtualModelManagement.tsx", 1_125],
  ["features/org-units/OrgUnitManagement.tsx", 1_200],
  ["features/quick-start/QuickStartManagement.tsx", 975],
  ["features/playground/PlaygroundPage.tsx", 950],
  ["features/prompts/PromptDetailPage.tsx", 900],
  ["features/teams/TeamManagement.tsx", 800],
  ["features/virtual-users/VirtualUserTokenManagement.tsx", 975],
] as const;

describe("client component performance boundaries", () => {
  it.each(componentBudgets)("keeps %s below %i lines", async (relativePath, maxLines) => {
    const source = await readSource(relativePath);
    const lineCount = source.split(/\r?\n/).length;

    expect(lineCount, `${relativePath} exceeded its orchestration-file budget`).toBeLessThanOrEqual(maxLines);
  });

  it("keeps the provider models editor behind a dynamic boundary", async () => {
    const source = await readSource("features/provider-accounts/ProviderAccountWizard.tsx");

    expect(source).toContain('import("@/features/provider-accounts/models-editor")');
    expect(source).not.toMatch(/from\s+["']@\/app\/components\/provider-account-wizard\/models-editor["']/);
  });

  it("keeps logs charts and log details out of the initial client graph", async () => {
    const source = await readSource("features/logs/LogsCommandCenter.tsx");

    expect(source).toContain('import("@/features/logs/logs-charts")');
    expect(source).toContain('import("@/features/logs/log-detail-sheet")');
    expect(source).not.toContain('from "recharts"');
  });

  it("keeps feature chart implementations out of their orchestration graphs", async () => {
    const [metrics, dashboard, billing] = await Promise.all([
      readSource("features/metrics/ModelMetricsDashboard.tsx"),
      readSource("features/dashboard/DashboardAnalytics.tsx"),
      readSource("features/billing/BillingFinancePanels.tsx"),
    ]);

    expect(metrics).toContain('import("@/features/metrics/metrics-charts")');
    expect(dashboard).toContain('import("@/features/dashboard/dashboard-charts")');
    expect(billing).toContain('import("@/features/billing/provider-cost-chart")');
    for (const source of [metrics, dashboard, billing]) {
      expect(source).not.toMatch(/from\s+["']recharts["']/u);
    }
  });
});

async function readSource(relativePath: string) {
  return readFile(path.join(appRoot, relativePath), "utf8");
}
