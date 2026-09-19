import { fireEvent, render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next/dynamic", () => ({
  default: () => function DynamicStub() {
    return null;
  },
}));

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    logs: (await import("@/messages/modules/logs/en.json")).default,
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

import { createTranslator } from "next-intl";

import logsMessages from "@/messages/modules/logs/en.json";

import { LogDetailSheet } from "./log-detail-sheet";
import type { LogDetail, RequestLog } from "./logs-types";
import { createLogFormatters } from "./logs-utils";

const t = createTranslator({ locale: "en", messages: { logs: logsMessages }, namespace: "logs" }) as unknown as ComponentProps<typeof LogDetailSheet>["t"];

const baseLog: RequestLog = {
  log_id: "log-1",
  request_id: "req-1",
  trace_id: "tr-1",
  model: "gpt-5",
  status_code: 200,
  started_at: "2026-09-19T10:00:00Z",
  duration_ms: 400,
  payload_mode: "metadata-only",
};

const spans = [
  { span_id: "root", name: "gateway.request", started_at: "2026-09-19T10:00:00.000Z", duration_ms: 400 },
  { span_id: "policy", parent_span_id: "root", name: "policy.evaluate", started_at: "2026-09-19T10:00:00.010Z", duration_ms: 40 },
  { span_id: "provider", parent_span_id: "root", name: "provider.call", started_at: "2026-09-19T10:00:00.060Z", duration_ms: 300 },
];

function detailFor(log: RequestLog): LogDetail {
  return { log, content_visibility: { can_view_content: false, request_available: false, response_available: false } };
}

function renderSheet(overrides: Partial<ComponentProps<typeof LogDetailSheet>> = {}) {
  const log = overrides.detail?.log ?? baseLog;
  return render(
    <LogDetailSheet
      selectedLog={log}
      detail={detailFor(log)}
      timingSpans={spans}
      traceHref="/traces?trace=tr-1"
      onFilterModel={vi.fn()}
      onClose={vi.fn()}
      formatters={createLogFormatters("en")}
      closeLabel="Close"
      t={t}
      {...overrides}
    />,
  );
}

describe("LogDetailSheet", () => {
  it("breaks the response time down by the trace's real spans", () => {
    renderSheet();

    const timing = screen.getByTestId("log-timing-breakdown");
    expect(timing).toHaveTextContent("policy.evaluate");
    expect(timing).toHaveTextContent("provider.call");
    expect(timing).toHaveTextContent("Total 400 ms");
  });

  it("renders no timing breakdown without a trace ID", () => {
    renderSheet({ detail: detailFor({ ...baseLog, trace_id: undefined }), traceHref: null });

    expect(screen.queryByTestId("log-timing-breakdown")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "View trace" })).not.toBeInTheDocument();
  });

  it("renders no timing breakdown for an unsampled trace", () => {
    renderSheet({ timingSpans: null });

    expect(screen.queryByTestId("log-timing-breakdown")).not.toBeInTheDocument();
  });

  it("shows the error block only for failed requests", () => {
    const { unmount } = renderSheet();
    expect(screen.queryByTestId("log-error-block")).not.toBeInTheDocument();
    unmount();

    renderSheet({ detail: detailFor({ ...baseLog, status_code: 502, error_type: "provider_timeout" }) });
    const block = screen.getByTestId("log-error-block");
    expect(block).toHaveTextContent("provider_timeout");
    expect(block).toHaveTextContent("The gateway or an upstream provider failed with HTTP 502.");
  });

  it("shows only the signals the log carries", () => {
    renderSheet({ detail: detailFor({ ...baseLog, retry_count: 2, cache_status: "miss" }) });

    expect(screen.getByText("Retry ×2")).toBeInTheDocument();
    expect(screen.queryByText("Cache")).not.toBeInTheDocument();
    expect(screen.queryByText("Fallback")).not.toBeInTheDocument();
  });

  it("filters on the model and links to the trace in the traces tab", () => {
    const onFilterModel = vi.fn();
    renderSheet({ onFilterModel });

    fireEvent.click(screen.getByRole("button", { name: "Filter on this model" }));
    expect(onFilterModel).toHaveBeenCalledWith("gpt-5");
    expect(screen.getByRole("link", { name: "View trace" })).toHaveAttribute("href", "/traces?trace=tr-1");
  });
});
