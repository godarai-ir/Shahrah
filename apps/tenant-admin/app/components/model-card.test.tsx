import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ModelCard, type ModelCardData } from "./model-card";

const model: ModelCardData = {
  id: "gpt-4.1-mini",
  name: "GPT-4.1 Mini",
  provider: {
    id: "openai",
    name: "OpenAI",
  },
  description: "Compact model",
  category: ["chat"],
  status: "stable",
  capabilities: ["chat"],
  contextWindow: 256_000,
  pricing: { inputPerMillion: 2, outputPerMillion: 8 },
  inputModalities: ["text"],
  outputModalities: ["text"],
  invocationModes: ["chat"],
  metadata: {
    modelId: "gpt-4.1-mini",
    mode: "chat",
  },
};

describe("ModelCard", () => {
  it("keeps account selection and comparison toggles independent", () => {
    const onSelectedChange = vi.fn();
    const onComparedChange = vi.fn();

    render(
      <ModelCard
        model={model}
        isSelected={false}
        onSelectedChange={onSelectedChange}
        selectedCheckboxLabel="Enabled"
        isCompared={false}
        onComparedChange={onComparedChange}
        compareCheckboxLabel="Compare"
      />,
    );

    fireEvent.click(screen.getByLabelText("Enabled"));
    expect(onSelectedChange).toHaveBeenCalledWith(true);
    expect(onComparedChange).not.toHaveBeenCalled();

    fireEvent.click(screen.getByLabelText("Compare"));
    expect(onComparedChange).toHaveBeenCalledWith(true);
  });

  it("can hide the copy action in focused workflows", () => {
    render(<ModelCard model={model} showCopyAction={false} />);

    expect(screen.queryByRole("button", { name: "Copy ID" })).not.toBeInTheDocument();
  });

  it("groups context, pricing, formats, capabilities, and invocation modes into readable panels", () => {
    render(<ModelCard model={model} />);

    expect(screen.getByRole("region", { name: "Context" })).toHaveTextContent("256K");
    expect(screen.getByRole("region", { name: "Input cost / Output cost" })).toHaveTextContent("$2 / 1M tokens");
    expect(screen.getByRole("region", { name: "Input cost / Output cost" })).toHaveTextContent("$8 / 1M tokens");
    expect(screen.getByRole("region", { name: "Input formats" })).toHaveTextContent("text");
    expect(screen.getByRole("region", { name: "Capabilities" })).toHaveTextContent("chat");
    expect(screen.getByRole("region", { name: "Invocation modes" })).toHaveTextContent("chat");
  });

  it("uses semantic tones to distinguish recommendation, formats, and capabilities", () => {
    render(
      <ModelCard
        model={{
          ...model,
          recommended: true,
          capabilities: ["function_calling", "cache_control", "structured_output"],
        }}
      />,
    );

    expect(screen.getByText("Recommended")).toHaveClass("bg-success/10");
    expect(within(screen.getByRole("region", { name: "Input formats" })).getByTitle("text")).toHaveClass("bg-info/10");
    expect(within(screen.getByRole("region", { name: "Output formats" })).getByTitle("text")).toHaveClass("bg-success/10");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("function_calling")).toHaveClass("bg-info/10");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("cache_control")).toHaveClass("bg-warning/10");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("structured_output")).toHaveClass("bg-success/10");
  });
});
