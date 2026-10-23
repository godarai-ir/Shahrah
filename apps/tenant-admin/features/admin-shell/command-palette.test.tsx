import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { PermissionsProvider } from "@/components/permissions-context";

import { CommandPaletteProvider } from "./command-palette-context";
import { RECENT_COMMANDS_STORAGE_KEY } from "./recent-commands";

const push = vi.fn();

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/",
  useRouter: () => ({ push }),
}));

beforeEach(() => {
  push.mockReset();
  window.localStorage.clear();
  // cmdk scrolls the selected item into view; jsdom does not implement it.
  Element.prototype.scrollIntoView = vi.fn();
});

function renderPalette(permissions: string[]) {
  return render(
    <PermissionsProvider permissions={permissions}>
      <CommandPaletteProvider>
        <input aria-label="page filter" />
      </CommandPaletteProvider>
    </PermissionsProvider>,
  );
}

function pressShortcut(target: Element | Window = window) {
  fireEvent.keyDown(target, { key: "k", ctrlKey: true });
}

describe("CommandPalette", () => {
  it("toggles with Ctrl+K, even while typing in an input", async () => {
    renderPalette([]);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    pressShortcut(screen.getByLabelText("page filter"));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();

    pressShortcut();
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("closes on Escape", async () => {
    renderPalette([]);
    pressShortcut();
    const dialog = await screen.findByRole("dialog");

    fireEvent.keyDown(dialog, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("ignores the shortcut when something else already handled it", () => {
    renderPalette([]);
    const event = new KeyboardEvent("keydown", { key: "k", ctrlKey: true, cancelable: true });
    event.preventDefault();
    window.dispatchEvent(event);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("navigates to the selected command and records it as recent", async () => {
    renderPalette([]);
    pressShortcut();
    const item = await screen.findByText("palette.actions.createSupportTicket");

    fireEvent.click(item);

    expect(push).toHaveBeenCalledWith("/support/tickets/new");
    expect(JSON.parse(window.localStorage.getItem(RECENT_COMMANDS_STORAGE_KEY) ?? "[]")).toEqual([
      "action:createSupportTicket",
    ]);
  });

  it("does not render commands the user lacks permission for", async () => {
    renderPalette([]);
    pressShortcut();
    await screen.findByRole("dialog");

    expect(screen.getByText("nav.dashboard")).toBeInTheDocument();
    expect(screen.queryByText("nav.logs")).not.toBeInTheDocument();
    expect(screen.queryByText("palette.actions.createProviderAccount")).not.toBeInTheDocument();
  });

  it("finds a Persian-titled page by its English route name", async () => {
    renderPalette(["logs:read_metadata"]);
    pressShortcut();
    const input = await screen.findByPlaceholderText("palette.placeholder");

    fireEvent.change(input, { target: { value: "LOGS" } });
    await waitFor(() => expect(screen.queryByText("nav.dashboard")).not.toBeInTheDocument());
    expect(screen.getByText("nav.logs")).toBeInTheDocument();
  });

  it("shows recent commands only while the query is empty", async () => {
    window.localStorage.setItem(RECENT_COMMANDS_STORAGE_KEY, JSON.stringify(["nav:quickStart", "nav:unknown"]));
    renderPalette([]);
    pressShortcut();

    expect(await screen.findByText("palette.groups.recent")).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText("palette.placeholder"), { target: { value: "quick" } });
    await waitFor(() => expect(screen.queryByText("palette.groups.recent")).not.toBeInTheDocument());
  });
});
