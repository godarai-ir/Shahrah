import { fireEvent, render, screen, within } from "@testing-library/react";
import type { ComponentProps } from "react";
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

vi.mock("@/features/logs/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./api")>();
  const pending = () => new Promise<never>(() => {});
  return {
    ...actual,
    getLogs: vi.fn(pending),
    getLogSummary: vi.fn(pending),
    getLogFacets: vi.fn(pending),
    getLogAnalytics: vi.fn(pending),
    getLogRangeStats: vi.fn(pending),
  };
});

const log = {
  log_id: "log-1",
  request_id: "req-1",
  provider: "openai",
  model: "gpt-5",
  status_code: 200,
  started_at: "2026-09-19T10:00:00Z",
  duration_ms: 320,
};

function renderCenter(overrides: Partial<ComponentProps<typeof LogsCommandCenter>> = {}) {
  return render(
    <LogsCommandCenter
      initialLogs={{ data: [log], total: 1, limit: 25, offset: 0 }}
      initialSummary={{ data: { request_count: 1, error_rate: 0 } }}
      initialFacets={{ data: {} }}
      initialSavedViews={{ data: [] }}
      initialTimeseries={{ data: [] }}
      initialBreakdown={{ data: [] }}
      initialRangeStats={{ data: {} }}
      {...overrides}
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

  it("shows saved views as chips above the table and highlights the applied one", () => {
    renderCenter({
      initialSavedViews: {
        data: [
          { saved_view_id: "v1", name: "OpenAI 7d", filters: { provider: "openai", range: "7d" }, visible_columns: [], shared_scope: "tenant" },
          { saved_view_id: "v2", name: "Legacy", filters: { from: "2026-01-01T00:00:00.000Z", to: "2026-01-02T00:00:00.000Z" }, visible_columns: [], shared_scope: "tenant" },
        ],
      },
    });

    const openai = screen.getByRole("button", { name: "OpenAI 7d" });
    const legacy = screen.getByRole("button", { name: "Legacy" });
    expect(openai.compareDocumentPosition(screen.getByRole("table")) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(openai).toHaveAttribute("aria-pressed", "false");
    // A legacy view that stored only derived timestamps matches the default 24h range.
    expect(legacy).toHaveAttribute("aria-pressed", "true");

    fireEvent.click(openai);

    expect(openai).toHaveAttribute("aria-pressed", "true");
    expect(legacy).toHaveAttribute("aria-pressed", "false");

    fireEvent.click(legacy);

    expect(legacy).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByRole("button", { name: /From:/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Saved views" })).not.toBeInTheDocument();
  });

  it("counts active filters on the filters button and hides the badge when there are none", () => {
    const empty = renderCenter();

    const toolbar = screen.getByRole("toolbar", { name: "Analysis Filters" });
    const toggle = within(toolbar).getByRole("button", { name: /^Filters/ });
    expect(within(toolbar).getByRole("button", { name: "Refresh" })).toBeInTheDocument();
    expect(within(toggle).queryByLabelText(/active filter/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Clear all" })).not.toBeInTheDocument();
    empty.unmount();

    renderCenter({
      initialSavedViews: {
        data: [{ saved_view_id: "v1", name: "Scoped", filters: { provider: "openai", status_code: "500" }, visible_columns: [], shared_scope: "tenant" }],
      },
    });
    fireEvent.click(screen.getByRole("button", { name: "Scoped" }));

    expect(screen.getByLabelText("2 active filters")).toHaveTextContent("2");
    expect(screen.getByRole("button", { name: "Clear all" })).toBeInTheDocument();
  });

  it("tints error rows and shows request signals", () => {
    renderCenter({
      initialLogs: {
        data: [
          { ...log, log_id: "ok", status_code: 200, cache_status: "hit" },
          { ...log, log_id: "failed", status_code: 502, retry_count: 2, fallback_count: 1 },
        ],
        total: 2,
        limit: 25,
        offset: 0,
      },
    });

    const [, okRow, failedRow] = screen.getAllByRole("row");
    expect(okRow).not.toHaveClass("bg-destructive/4");
    expect(failedRow).toHaveClass("bg-destructive/4");
    expect(within(okRow).getByText("Cache")).toBeInTheDocument();
    expect(within(failedRow).getByText("Retry ×2")).toBeInTheDocument();
    expect(within(failedRow).getByText("Fallback")).toBeInTheDocument();
  });

  it("keeps metric cards and charts out of the page until the summary is opened", () => {
    renderCenter();

    expect(screen.queryByText("P95 latency")).not.toBeInTheDocument();
    expect(screen.queryByText("Traffic and Errors")).not.toBeInTheDocument();
  });
});
