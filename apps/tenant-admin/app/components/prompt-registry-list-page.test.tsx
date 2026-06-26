import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PromptRegistryListPage from "./prompt-registry-list-page";

const pushMock = vi.fn();
const writeTextMock = vi.fn();
const confirmMock = vi.fn();
const requestJSONMock = vi.fn();
const tMock = (key: string, values?: Record<string, unknown>) => {
  if (!values) return key;
  return Object.entries(values).reduce((text, [name, value]) => text.replace(`{${name}}`, String(value)), key);
};

vi.mock("next-intl", () => ({
  useLocale: () => "en",
  useTranslations: () => tMock,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

vi.mock("@/app/components/provider-account-wizard/utils", () => ({
  normalizeError: (error: unknown, fallback: string) => (error instanceof Error ? error.message : fallback),
  requestJSON: (path: string, options?: RequestInit) => requestJSONMock(path, options),
}));

describe("PromptRegistryListPage", () => {
  beforeEach(() => {
    pushMock.mockReset();
    writeTextMock.mockReset();
    confirmMock.mockReset();
    requestJSONMock.mockReset();
    writeTextMock.mockResolvedValue(undefined);
    confirmMock.mockReturnValue(true);
    vi.stubGlobal("navigator", {
      clipboard: {
        writeText: writeTextMock,
      },
    });
    vi.stubGlobal("confirm", confirmMock);
    requestJSONMock.mockImplementation(async (path: string, options?: RequestInit) => {
      if (path.startsWith("/api/prompts?")) {
        return {
          data: [
            {
              id: "prompt-1",
              name: "booking_greeting",
              slug: "booking_greeting",
              description: "Greets a booking customer",
              latest_version: 3,
              latest_version_detail: { model: "gpt-4.1-mini", fqn: "booking_greeting.v3" },
              updated_at: "2026-06-26T08:00:00Z",
              created_by: "user-1",
            },
          ],
          total: 1,
        };
      }
      if (path === "/api/gateway/lookups/users") {
        return {
          data: [
            {
              id: "user-1",
              display_name: "Jane Doe",
              email: "jane@example.com",
            },
          ],
          total: 1,
        };
      }
      if (path === "/api/prompts/prompt-1" && options?.method === "DELETE") {
        return { ok: true };
      }
      return { data: [], total: 0 };
    });
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the list page without builder UI", async () => {
    render(<PromptRegistryListPage />);

    expect(await screen.findByText("booking_greeting")).toBeInTheDocument();
    expect(screen.queryByText("playground.title")).not.toBeInTheDocument();
    expect(screen.queryByText("builder.titleNew")).not.toBeInTheDocument();
  });

  it("navigates to the create page from the CTA", async () => {
    render(<PromptRegistryListPage />);

    await screen.findByText("booking_greeting");
    fireEvent.click(screen.getByText("actions.createPrompt"));

    expect(pushMock).toHaveBeenCalledWith("/prompts/new");
  });

  it("navigates to the prompt detail page when a row is clicked", async () => {
    render(<PromptRegistryListPage />);

    fireEvent.click(await screen.findByText("booking_greeting"));

    await waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith("/prompts/prompt-1");
    });
  });

  it("shows the creator display name instead of the raw user id", async () => {
    render(<PromptRegistryListPage />);

    expect(await screen.findByText("Jane Doe")).toBeInTheDocument();
    expect(screen.queryByText("user-1")).not.toBeInTheDocument();
  });

  it("renders a row action menu trigger", async () => {
    render(<PromptRegistryListPage />);

    await screen.findByText("Jane Doe");
    expect(screen.getByRole("button", { name: "actions.openMenu" })).toBeInTheDocument();
  });
});
