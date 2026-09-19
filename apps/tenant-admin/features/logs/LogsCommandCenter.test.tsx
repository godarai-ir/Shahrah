import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import LogsCommandCenter from "./LogsCommandCenter";

vi.mock("next/navigation", () => ({
  usePathname: () => "/logs",
  useRouter: () => ({ replace: vi.fn(), push: vi.fn(), refresh: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("next/dynamic", () => ({
  default: () => function DynamicStub() {
    return null;
  },
}));

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    logs: (await import("@/messages/modules/logs/en.json")).default,
    common: (await import("@/messages/modules/common/en.json")).default,
    settings: (await import("@/messages/modules/settings/en.json")).default,
    providerAccounts: (await import("@/messages/modules/providerAccounts/en.json")).default,
  };
  const translators = new Map<string, unknown>();
  return {
    ...actual,
    useLocale: () => "en",
    useTranslations: (namespace: string) => {
      if (!translators.has(namespace)) translators.set(namespace, actual.createTranslator({ locale: "en", messages, namespace: namespace as never }));
      return translators.get(namespace);
    },
  };
});

vi.mock("@/features/logs/RequestLoggingSettingsPanel", () => ({ default: () => null }));

const log = {
  log_id: "log-1",
  request_id: "req-1",
  provider: "openai",
  model: "gpt-5",
  status_code: 200,
  started_at: "2026-09-19T10:00:00Z",
  duration_ms: 320,
};

function renderCenter() {
  return render(
    <LogsCommandCenter
      initialLogs={{ data: [log], total: 1, limit: 25, offset: 0 }}
      initialSummary={{ data: { request_count: 1, error_rate: 0 } }}
      initialFacets={{ data: {} }}
      initialSavedViews={{ data: [] }}
      initialTimeseries={{ data: [] }}
      initialBreakdown={{ data: [] }}
      initialRangeStats={{ data: {} }}
    />,
  );
}

describe("LogsCommandCenter layout", () => {
  it("puts the filter bar, then the collapsed summary, then the table", () => {
    const { container } = renderCenter();

    const range = screen.getByRole("combobox", { name: "Range" });
    const summary = screen.getByTestId("logs-range-summary");
    const table = screen.getByRole("table");

    expect(range.compareDocumentPosition(summary) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(summary.compareDocumentPosition(table) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(summary).not.toHaveAttribute("open");
    expect(container.querySelectorAll("details")).toHaveLength(1);
  });

  it("keeps metric cards and charts out of the page until the summary is opened", () => {
    renderCenter();

    expect(screen.queryByText("P95 latency")).not.toBeInTheDocument();
    expect(screen.queryByText("Traffic and Errors")).not.toBeInTheDocument();
  });
});
