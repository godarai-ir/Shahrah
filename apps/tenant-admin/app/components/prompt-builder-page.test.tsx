import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PromptBuilderPage from "./prompt-builder-page";

const pushMock = vi.fn();

vi.mock("next-intl", () => ({
  useLocale: () => "en",
  useTranslations: () => (key: string, values?: Record<string, unknown>) => {
    if (!values) return key;
    return Object.entries(values).reduce((text, [name, value]) => text.replace(`{${name}}`, String(value)), key);
  },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

describe("PromptBuilderPage", () => {
  beforeEach(() => {
    pushMock.mockReset();
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("creates a prompt on the dedicated create page", async () => {
    vi.stubGlobal("fetch", vi.fn(createFetchMock()));

    render(<PromptBuilderPage mode="create" />);

    fireEvent.click(await screen.findByText("actions.selectModel"));
    fireEvent.click(await screen.findByText("GPT 4.1 Mini"));
    fireEvent.change(screen.getByPlaceholderText("builder.fields.namePlaceholder"), { target: { value: "booking_greeting" } });
    fireEvent.click(screen.getByText("actions.saveNewPrompt"));

    await waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith("/prompts/prompt-1");
    });
  }, 10000);

  it("saves a new immutable version on the dedicated edit page", async () => {
    const fetchMock = vi.fn(createFetchMock({ editMode: true }));
    vi.stubGlobal("fetch", fetchMock);

    render(<PromptBuilderPage mode="edit" promptID="prompt-1" />);

    expect(await screen.findByDisplayValue("booking_greeting")).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText("builder.fields.commitMessagePlaceholder"), { target: { value: "Refined wording" } });
    fireEvent.click(screen.getByText("actions.saveNewVersion"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/prompts/prompt-1/versions",
        expect.objectContaining({
          method: "POST",
          cache: "no-store",
        }),
      );
    });
  });

  it("renders playground responses in an LTR highlighted code block", async () => {
    vi.stubGlobal("fetch", vi.fn(createFetchMock({ runMode: true })));

    render(<PromptBuilderPage mode="create" />);

    fireEvent.click(await screen.findByText("actions.selectModel"));
    fireEvent.click(await screen.findByText("GPT 4.1 Mini"));
    fireEvent.change(screen.getByPlaceholderText("builder.fields.namePlaceholder"), { target: { value: "booking_greeting" } });
    fireEvent.click(screen.getByText("actions.run"));

    await waitFor(() => {
      const codeBlock = document.querySelector("pre");
      expect(codeBlock).not.toBeNull();
      expect(codeBlock).toHaveAttribute("dir", "ltr");
      expect(codeBlock?.textContent).toContain("Hello from Shahrah playground");
    });
  });
});

function createFetchMock(options?: { editMode?: boolean; runMode?: boolean }) {
  let savedVersion = false;

  return async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);

    if (url === "/api/models") {
      return jsonResponse({
        data: [
          {
            id: "gpt-4.1-mini",
            object: "model",
            owned_by: "openai",
            provider: "openai",
            aliases: ["gpt-4.1-mini"],
          },
        ],
      });
    }
    if (url === "/api/provider-catalog/providers") {
      return jsonResponse({
        data: [{ provider: "openai", display_name: "OpenAI", description: "Primary models" }],
      });
    }
    if (url === "/api/provider-catalog/providers/openai/models") {
      return jsonResponse({
        data: [
          {
            provider: "openai",
            model: "gpt-4.1-mini",
            display_name: "GPT 4.1 Mini",
            capabilities: ["chat"],
            context_window: 128000,
          },
        ],
      });
    }
    if (url === "/api/guardrails?limit=200") {
      return jsonResponse({ data: [] });
    }
    if (options?.editMode && url === "/api/prompts/prompt-1") {
      return jsonResponse(buildPromptDetail(savedVersion ? "version-2" : "version-1"));
    }
    if (options?.editMode && url === "/api/prompts/prompt-1/versions?limit=100") {
      return jsonResponse({
        data: savedVersion
          ? [buildPromptVersion("version-2", 2), buildPromptVersion("version-1", 1)]
          : [buildPromptVersion("version-1", 1)],
      });
    }
    if (url === "/api/prompts" && init?.method === "POST") {
      return jsonResponse(buildPromptVersion("version-1", 1));
    }
    if (url === "/api/prompts/prompt-1/versions" && init?.method === "POST") {
      savedVersion = true;
      return jsonResponse(buildPromptVersion("version-2", 2));
    }
    if (url === "/api/prompts/run" && init?.method === "POST" && options?.runMode) {
      return jsonResponse({
        choices: [{ message: { role: "assistant", content: "Hello from Shahrah playground" } }],
        usage: { prompt_tokens: 10, completion_tokens: 6, total_tokens: 16 },
      });
    }

    return jsonResponse({ data: [] });
  };
}

function buildPromptDetail(versionID: string) {
  return {
    data: {
      id: "prompt-1",
      name: "booking_greeting",
      slug: "booking_greeting",
      description: "Greets a booking customer",
      latest_version: versionID === "version-2" ? 2 : 1,
      latest_version_detail: buildPromptVersion(versionID, versionID === "version-2" ? 2 : 1),
      updated_at: "2026-06-26T08:00:00Z",
      created_by: "user-1",
    },
  };
}

function buildPromptVersion(id: string, version: number) {
  return {
    id,
    prompt_id: "prompt-1",
    version,
    fqn: `prompts.booking_greeting.v${version}`,
    name: "booking_greeting",
    description: "Greets a booking customer",
    model: "gpt-4.1-mini",
    messages: [{ role: "system", content: "You are a helpful assistant." }],
    manifest: {
      type: "chat_prompt",
      name: "booking_greeting",
      description: "Greets a booking customer",
      model: "gpt-4.1-mini",
      messages: [{ role: "system", content: "You are a helpful assistant." }],
      variables: {},
      parameters: { stream: true, temperature: 0.2 },
      model_settings: { stream: true, temperature: 0.2 },
      metadata: {},
      cache: { enabled: false },
      guardrails: { input: [], output: [], scope: "last" },
    },
  };
}

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
