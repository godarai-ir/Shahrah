import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { RequestTrace } from "./trace-types";

const replace = vi.fn();
vi.mock("next/navigation", () => ({
  usePathname: () => "/traces",
  useRouter: () => ({ replace, push: vi.fn(), refresh: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("next/dynamic", () => ({
  default: () => function DynamicStub() {
    return null;
  },
}));

vi.mock("@/hooks/use-mobile", () => ({ useIsMobile: () => false }));

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    traces: (await import("@/messages/modules/traces/en.json")).default,
    common: (await import("@/messages/modules/common/en.json")).default,
    settings: (await import("@/messages/modules/settings/en.json")).default,
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

const sampled: RequestTrace = {
  trace_id: "tr_sampled",
  tenant_id: "t1",
  request_id: "req_1",
  model: "gpt-5",
  provider: "openai",
  route: "/v1/chat/completions",
  status_code: 200,
  started_at: "2026-09-19T10:00:00Z",
  duration_ms: 420,
  sampled: true,
  attempts: 2,
};
const unsampled: RequestTrace = { ...sampled, trace_id: "tr_unsampled", sampled: false, attempts: 1 };

const api = vi.hoisted(() => ({
  loadTraces: vi.fn(),
  getTraceDetail: vi.fn(),
  getTraceTimeline: vi.fn(),
  getTraceAttempts: vi.fn(),
}));

vi.mock("@/features/traces/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./api")>()),
  ...api,
}));

import TraceManagement from "./TraceManagement";

function renderTraces() {
  return render(
    <TraceManagement
      initialData={[sampled, unsampled]}
      initialTotal={2}
      initialError={null}
      initialTracePrivacySettings={{ capture_policy: "metadata-only" }}
      initialTracePrivacyError={null}
      currentUserID="u1"
      currentUserLabel="u1@example.com"
    />,
  );
}

describe("TraceManagement master–detail", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.loadTraces.mockResolvedValue({ list: { data: [sampled, unsampled], total: 2 }, facets: { data: {} } });
    api.getTraceDetail.mockImplementation(async (id: string) => ({
      trace: { ...(id === sampled.trace_id ? sampled : unsampled), attempts: 1 },
      summary: { spans_count: 2, attempts_count: 1, errors_count: 0 },
    }));
    api.getTraceAttempts.mockResolvedValue([]);
    api.getTraceTimeline.mockResolvedValue([
      { span_id: "s1", trace_id: sampled.trace_id, name: "gateway.request", kind: "server", started_at: "2026-09-19T10:00:00.000Z", duration_ms: 420, status: "ok", attributes: {} },
      { span_id: "s2", trace_id: sampled.trace_id, name: "provider.call", kind: "client", started_at: "2026-09-19T10:00:00.050Z", duration_ms: 350, status: "ok", attributes: {} },
    ]);
  });

  it("shows attempts and model on list rows and waits for a selection before loading a timeline", async () => {
    renderTraces();

    const row = screen.getByRole("button", { name: "Show trace tr_sampled" });
    expect(row).toHaveTextContent("gpt-5");
    expect(row).toHaveTextContent("2 attempts");
    expect(screen.getByText("Select a trace")).toBeInTheDocument();
    await waitFor(() => expect(api.loadTraces).toHaveBeenCalled());
    expect(api.getTraceTimeline).not.toHaveBeenCalled();
  });

  it("loads the selected trace's timeline once", async () => {
    renderTraces();

    fireEvent.click(screen.getByRole("button", { name: "Show trace tr_sampled" }));

    expect(await screen.findByText("provider.call")).toBeInTheDocument();
    expect(screen.getByText("2 span(s)")).toBeInTheDocument();
    await waitFor(() => expect(api.loadTraces).toHaveBeenCalled());
    expect(api.getTraceTimeline).toHaveBeenCalledTimes(1);
    expect(api.getTraceTimeline).toHaveBeenCalledWith("tr_sampled", expect.any(AbortSignal));
    expect(screen.getByRole("button", { name: "Show trace tr_sampled" })).toHaveAttribute("aria-pressed", "true");
  });

  it("shows an explicit empty state for an unsampled trace instead of a lasting spinner", async () => {
    renderTraces();

    fireEvent.click(screen.getByRole("button", { name: "Show trace tr_unsampled" }));

    expect(await screen.findByText("This request was not sampled — only the summary is available.")).toBeInTheDocument();
    expect(screen.getByText("Current capture policy: Metadata-only")).toBeInTheDocument();
    expect(document.querySelector("[aria-busy='true']")).toBeNull();
    expect(api.getTraceTimeline).not.toHaveBeenCalled();
  });

  it("aborts the previous trace's requests when the selection changes", async () => {
    renderTraces();

    fireEvent.click(screen.getByRole("button", { name: "Show trace tr_sampled" }));
    const firstSignal = api.getTraceDetail.mock.calls[0][1] as AbortSignal;
    fireEvent.click(screen.getByRole("button", { name: "Show trace tr_unsampled" }));

    expect(firstSignal.aborted).toBe(true);
    expect(await screen.findByText(/not sampled/)).toBeInTheDocument();
  });

  it("keeps the selection in the URL", async () => {
    renderTraces();

    fireEvent.click(screen.getByRole("button", { name: "Show trace tr_sampled" }));

    await waitFor(() => expect(replace).toHaveBeenLastCalledWith("/traces?trace=tr_sampled", { scroll: false }));
  });
});
