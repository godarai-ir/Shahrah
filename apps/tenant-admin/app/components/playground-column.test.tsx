import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createPlaygroundColumn, type PlaygroundRunResult } from "./playground-domain";
import { PlaygroundOutputPanel } from "./playground-column";

describe("playground image output", () => {
  beforeEach(() => {
    vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => window.setTimeout(() => callback(0), 0));
    vi.stubGlobal("cancelAnimationFrame", (handle: number) => window.clearTimeout(handle));
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: vi.fn(() => "blob:generated-image") });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: vi.fn() });
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders a base64 result through a revocable object URL", async () => {
    const column = createPlaygroundColumn("Option 1", "System");
    const result: PlaygroundRunResult = {
      status: "success",
      responseText: "",
      rawResponse: { data: [{ b64_json: "[base64 omitted]" }] },
      toolCalls: [],
      generatedImages: [{ index: 0, b64Json: "YWJj", mimeType: "image/png", revisedPrompt: "A revised prompt" }],
      generatedVideos: [],
    };
    const rendered = render(
      <PlaygroundOutputPanel
        column={column}
        mode="image_generation"
        result={result}
        locale="en"
        readOnly={false}
        canSave={false}
        canLoad={false}
        canOpenTrace={false}
        comparison={false}
        onChange={vi.fn()}
        onRun={vi.fn()}
        onStop={vi.fn()}
        onOpenModel={vi.fn()}
        onOpenSettings={vi.fn()}
        onOpenSave={vi.fn()}
        onOpenLoad={vi.fn()}
        t={(key, values) => key === "image.alt" ? `Generated image ${values?.count}` : key}
      />,
    );

    await waitFor(() => expect(screen.getByRole("img", { name: "Generated image 1" })).toHaveAttribute("src", "blob:generated-image"));
    expect(URL.createObjectURL).toHaveBeenCalledOnce();

    rendered.unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:generated-image");
  });
});
