import { fireEvent, render, screen } from "@testing-library/react";
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
});
