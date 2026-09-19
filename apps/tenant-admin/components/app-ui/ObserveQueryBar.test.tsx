import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = { common: (await import("@/messages/modules/common/en.json")).default };
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

import { EMPTY_OBSERVE_FILTERS } from "@/lib/observe-filters";

import { ObserveQueryBar } from "./ObserveQueryBar";

describe("ObserveQueryBar", () => {
  it("marks a shared filter the view cannot apply instead of dropping it", () => {
    render(<ObserveQueryBar tab="logs" filters={{ ...EMPTY_OBSERVE_FILTERS, errorCode: "timeout", model: "gpt-5" }} onFiltersChange={vi.fn()} />);

    const errorChip = screen.getByText("Error code: timeout").closest("[data-applied]");
    const modelChip = screen.getByText("Model: gpt-5").closest("[data-applied]");
    expect(errorChip).toHaveAttribute("data-applied", "false");
    expect(errorChip).toHaveTextContent("Not applied in this view");
    expect(modelChip).toHaveAttribute("data-applied", "true");
  });

  it("applies the same error code on traces", () => {
    render(<ObserveQueryBar tab="traces" filters={{ ...EMPTY_OBSERVE_FILTERS, errorCode: "timeout" }} onFiltersChange={vi.fn()} />);

    expect(screen.getByText("Error code: timeout").closest("[data-applied]")).toHaveAttribute("data-applied", "true");
    expect(screen.queryByText(/Not applied/)).not.toBeInTheDocument();
  });

  it("removes one chip and clears everything including view filters", () => {
    const onFiltersChange = vi.fn();
    const onClearAll = vi.fn();
    render(
      <ObserveQueryBar
        tab="logs"
        filters={{ ...EMPTY_OBSERVE_FILTERS, status: "500", provider: "openai" }}
        onFiltersChange={onFiltersChange}
        extraChips={[{ key: "route", label: "Route: /v1/chat" }]}
        onClearAll={onClearAll}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Remove Status filter" }));
    expect(onFiltersChange).toHaveBeenLastCalledWith({ ...EMPTY_OBSERVE_FILTERS, provider: "openai" });

    fireEvent.click(screen.getByRole("button", { name: "Clear all" }));
    expect(onClearAll).toHaveBeenCalledTimes(1);
    expect(onFiltersChange).toHaveBeenLastCalledWith(EMPTY_OBSERVE_FILTERS);
  });

  it("says the filters are shared and hides clear-all without filters", () => {
    render(<ObserveQueryBar tab="metrics" filters={EMPTY_OBSERVE_FILTERS} onFiltersChange={vi.fn()} />);

    expect(screen.getByText("Search, range and filters are shared across all three views.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Clear all" })).not.toBeInTheDocument();
  });
});
