import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PromptDetailPage from "./prompt-detail-page";
import { buildGatewaySnippet, buildSdkSnippet, serializePromptSpec } from "./prompt-detail-utils";

const pushMock = vi.fn();
const writeTextMock = vi.fn();
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
  requestJSON: (path: string) => requestJSONMock(path),
  unwrapData: <T,>(payload: { data?: T } | T) => (payload && typeof payload === "object" && "data" in payload ? payload.data ?? null : payload),
}));

describe("PromptDetailPage", () => {
  beforeEach(() => {
    pushMock.mockReset();
    requestJSONMock.mockReset();
    writeTextMock.mockReset();
    writeTextMock.mockResolvedValue(undefined);
    vi.stubGlobal("navigator", {
      clipboard: {
        writeText: writeTextMock,
      },
    });
    requestJSONMock.mockImplementation(async (path: string) => {
      if (path === "/api/prompts/prompt-1") {
        return {
          data: {
            id: "prompt-1",
            name: "booking_greeting",
            slug: "booking_greeting",
            description: "Greets a booking customer",
            latest_version: 2,
            latest_version_detail: buildPromptVersion("version-2", 2),
            updated_at: "2026-06-26T08:00:00Z",
            created_by: "user-2",
          },
        };
      }
      if (path === "/api/prompts/prompt-1/versions?limit=100") {
        return {
          data: [buildPromptVersion("version-2", 2), buildPromptVersion("version-1", 1)],
        };
      }
      if (path === "/api/gateway/lookups/users") {
        return {
          data: [
            { id: "user-1", display_name: "Old Owner" },
            { id: "user-2", display_name: "Latest Owner" },
          ],
        };
      }
      return { data: [] };
    });
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders a read-only prompt detail page by default", async () => {
    render(<PromptDetailPage promptID="prompt-1" />);

    expect(await screen.findByRole("heading", { name: "booking_greeting" })).toBeInTheDocument();
    expect(screen.getAllByText("Latest Owner").length).toBeGreaterThan(0);
    expect(screen.queryByPlaceholderText("builder.fields.namePlaceholder")).not.toBeInTheDocument();
    expect(screen.getByText("Latest system prompt")).toBeInTheDocument();
  });

  it("switches selected versions from history", async () => {
    render(<PromptDetailPage promptID="prompt-1" />);

    await screen.findByText("Latest system prompt");
    fireEvent.click(screen.getByText("actions.versionHistory"));
    fireEvent.click(await screen.findByText("Initial draft"));

    await waitFor(() => {
      expect(screen.getByText("Old system prompt")).toBeInTheDocument();
    });
  });

  it("builds the spec, sdk, and gateway snippets exposed by the code tabs", async () => {
    const version = buildPromptVersion("version-2", 2);
    expect(serializePromptSpec(version)).toContain('"name": "booking_greeting"');
    expect(buildSdkSnippet(version)).toContain("from openai import OpenAI");
    expect(buildGatewaySnippet(version, false)).toContain("curl --location");

    render(<PromptDetailPage promptID="prompt-1" />);

    await screen.findByText("Latest system prompt");

    expect(screen.getByRole("tab", { name: "detail.tabs.spec" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "detail.tabs.sdk" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "detail.tabs.gateway" })).toBeInTheDocument();
  });

  it("copies the selected FQN and opens compare dialog", async () => {
    render(<PromptDetailPage promptID="prompt-1" />);

    await screen.findByText("Latest system prompt");
    fireEvent.click(screen.getByText("actions.copyFqn"));
    await waitFor(() => {
      expect(writeTextMock).toHaveBeenCalledWith("prompts.booking_greeting.v2");
    });

    fireEvent.click(screen.getByText("actions.versionHistory"));
    fireEvent.click(await screen.findByText("actions.compareVersions"));
    expect(await screen.findByText("detail.compareTitle")).toBeInTheDocument();
    expect(
      screen.getAllByText((_, node) => node?.textContent?.includes('"description": "Latest greeting copy"') ?? false).length,
    ).toBeGreaterThan(0);
    expect(
      screen.getAllByText((_, node) => node?.textContent?.includes('"description": "Initial greeting copy"') ?? false).length,
    ).toBeGreaterThan(0);
  });

  it("loads the selected prompt version in Playground when creating a version", async () => {
    render(<PromptDetailPage promptID="prompt-1" />);

    await screen.findByText("Latest system prompt");
    expect(screen.queryByText("actions.editLatestVersion")).not.toBeInTheDocument();

    fireEvent.click(screen.getByText("actions.createVersion"));
    expect(pushMock).toHaveBeenCalledWith("/playground?promptID=prompt-1&versionID=version-2");

    fireEvent.click(screen.getByText("actions.tryPrompt"));
    expect(pushMock).toHaveBeenCalledWith("/playground?promptID=prompt-1&versionID=version-2");
  });
});

function buildPromptVersion(id: string, version: number) {
  const isLatest = version === 2;
  return {
    id,
    prompt_id: "prompt-1",
    version,
    fqn: `prompts.booking_greeting.v${version}`,
    name: "booking_greeting",
    description: isLatest ? "Latest greeting copy" : "Initial greeting copy",
    model: "gpt-4.1-mini",
    messages: [
      { role: "system", content: isLatest ? "Latest system prompt" : "Old system prompt" },
      { role: "user", content: isLatest ? "Hello {{customer_name}}" : "Hello there" },
    ],
    manifest: {
      type: "chat_prompt",
      name: "booking_greeting",
      description: isLatest ? "Latest greeting copy" : "Initial greeting copy",
      model: "gpt-4.1-mini",
      messages: [
        { role: "system", content: isLatest ? "Latest system prompt" : "Old system prompt" },
        { role: "user", content: isLatest ? "Hello {{customer_name}}" : "Hello there" },
      ],
      variables: isLatest ? { customer_name: "" } : {},
      parameters: { stream: !isLatest, temperature: isLatest ? 0.2 : 0.6 },
      model_settings: { stream: !isLatest, temperature: isLatest ? 0.2 : 0.6 },
      metadata: isLatest ? { audience: "customer" } : {},
      cache: { enabled: isLatest, type: "simple", ttl: 300 },
      guardrails: { input: isLatest ? ["guardrails/pii"] : [], output: [], scope: "last" },
    },
    commit_message: isLatest ? "Latest draft" : "Initial draft",
    created_by: isLatest ? "user-2" : "user-1",
    created_at: "2026-06-26T08:00:00Z",
  };
}
