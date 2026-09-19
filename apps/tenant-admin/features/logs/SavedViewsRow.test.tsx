import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { createTranslator } from "next-intl";
import { describe, expect, it, vi } from "vitest";

import messages from "@/messages/modules/logs/en.json";

import type { SavedView } from "./logs-types";
import { SavedViewsRow } from "./SavedViewsRow";

const t = createTranslator({ locale: "en", messages: { logs: messages }, namespace: "logs" }) as never;

const views: SavedView[] = [
  { saved_view_id: "v1", name: "Errors only", filters: { status_code: "500" }, visible_columns: [], shared_scope: "tenant" },
  { saved_view_id: "v2", name: "OpenAI", filters: { provider: "openai" }, visible_columns: [], shared_scope: "tenant" },
];

function renderRow(overrides: Partial<Parameters<typeof SavedViewsRow>[0]> = {}) {
  const props = {
    views,
    activeViewID: "v2",
    onApply: vi.fn(),
    onDelete: vi.fn(),
    onSaveCurrent: vi.fn(),
    deletingViewID: null,
    t,
    ...overrides,
  };
  render(<SavedViewsRow {...props} />);
  return props;
}

describe("SavedViewsRow", () => {
  it("marks only the view matching the current filters as pressed", () => {
    renderRow();

    expect(screen.getByRole("button", { name: "OpenAI" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Errors only" })).toHaveAttribute("aria-pressed", "false");
  });

  it("presses nothing when no view matches", () => {
    renderRow({ activeViewID: null });

    for (const view of views) {
      expect(screen.getByRole("button", { name: view.name })).toHaveAttribute("aria-pressed", "false");
    }
  });

  it("applies the clicked view", () => {
    const props = renderRow();

    fireEvent.click(screen.getByRole("button", { name: "Errors only" }));

    expect(props.onApply).toHaveBeenCalledWith(views[0]);
  });

  it("offers only the save chip when there are no views", () => {
    const props = renderRow({ views: [], activeViewID: null });

    expect(screen.getAllByRole("button")).toHaveLength(1);
    fireEvent.click(screen.getByRole("button", { name: messages.savedViews.saveCurrent }));
    expect(props.onSaveCurrent).toHaveBeenCalledTimes(1);
  });

  it("deletes a view only after confirmation", async () => {
    const props = renderRow();

    fireEvent.click(screen.getByRole("button", { name: "Delete view OpenAI" }));
    expect(props.onDelete).not.toHaveBeenCalled();

    fireEvent.click(await screen.findByRole("button", { name: messages.actions.delete }));
    await waitFor(() => expect(props.onDelete).toHaveBeenCalledWith("v2"));
  });
});
