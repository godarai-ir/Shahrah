import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ModelCard, type ModelCardData } from "./ModelCard";

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

  it("places selection in the title row and tags directly below the title", () => {
    const { container } = render(
      <ModelCard
        model={{ ...model, recommended: true }}
        isSelected={false}
        onSelectedChange={vi.fn()}
        selectedCheckboxLabel="Enabled"
        showCopyAction={false}
      />,
    );

    const title = screen.getByRole("heading", { name: "GPT-4.1 Mini" });
    const titleRow = container.querySelector<HTMLElement>('[data-slot="model-card-title-row"]');
    const tagRow = container.querySelector<HTMLElement>('[data-slot="model-card-tag-row"]');

    expect(title).toHaveClass("text-xl");
    expect(titleRow).toContainElement(screen.getByLabelText("Enabled"));
    expect(titleRow).toContainElement(tagRow);
    expect(within(tagRow as HTMLElement).getByText("Chat")).toBeInTheDocument();
    expect(within(tagRow as HTMLElement).getByText("Recommended")).toBeInTheDocument();
  });

  it("keeps the visible selection label concise while preserving the accessible model context", () => {
    render(
      <ModelCard
        model={model}
        isSelected
        onSelectedChange={vi.fn()}
        selectedCheckboxLabel="Enabled GPT-4.1 Mini"
        selectedControlLabel="Enabled"
        showCopyAction={false}
      />,
    );

    expect(screen.getByText("Enabled")).toBeVisible();
    expect(screen.getByLabelText("Enabled GPT-4.1 Mini")).toBeChecked();
    expect(screen.queryByText("Enabled GPT-4.1 Mini")).not.toBeInTheDocument();
  });

  it("groups context, pricing, formats, capabilities, and invocation modes into readable panels", () => {
    render(<ModelCard model={model} />);

    expect(screen.getByRole("region", { name: "Context" })).toHaveTextContent("256K");
    const pricingRegion = screen.getByRole("region", {
      name: "Cost per 1M tokens (some models use tiered pricing)",
    });
    expect(pricingRegion).toHaveTextContent("$2");
    expect(pricingRegion).toHaveTextContent("$8");
    expect(pricingRegion).not.toHaveTextContent("/ 1M tokens");
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

    expect(screen.getByText("Recommended")).toHaveClass("bg-success-soft");
    expect(within(screen.getByRole("region", { name: "Input formats" })).getByTitle("text")).toHaveClass("bg-info-soft");
    expect(within(screen.getByRole("region", { name: "Output formats" })).getByTitle("text")).toHaveClass("bg-success-soft");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("function_calling")).toHaveClass("bg-info-soft");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("cache_control")).toHaveClass("bg-warning-soft");
    expect(within(screen.getByRole("region", { name: "Capabilities" })).getByTitle("structured_output")).toHaveClass("bg-success-soft");
  });
});
