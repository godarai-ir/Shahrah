import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { getVariableColorStyle, VariableTextarea } from "./VariableTextarea";

describe("VariableTextarea", () => {
  it("draws its focus ring inside the overlay so scroll containers cannot clip its sides", () => {
    render(<VariableTextarea aria-label="Prompt" variables={[]} />);

    const textarea = screen.getByRole("textbox", { name: "Prompt" });
    const overlay = textarea.nextElementSibling;

    expect(textarea).toHaveClass("peer", "focus-visible:!outline-none", "focus-visible:!ring-0");
    expect(overlay).toHaveClass("peer-focus-visible:ring-[3px]", "peer-focus-visible:ring-inset", "peer-focus-visible:ring-ring/50");
  });
});

describe("getVariableColorStyle", () => {
  it("uses one color for the same variable and distinct colors for different variables", () => {
    const variables = ["context", "quest", "tone", ...Array.from({ length: 21 }, (_, index) => `variable_${index}`)];

    expect(getVariableColorStyle("context", variables)).toEqual(getVariableColorStyle("context", variables));
    expect(getVariableColorStyle("context", variables)).not.toEqual(getVariableColorStyle("quest", variables));
    expect(new Set(variables.map((name) => getVariableColorStyle(name, variables)["--variable-hue"])).size).toBe(variables.length);
  });

  it("keeps a variable color stable while its name is being typed in the same slot", () => {
    expect(getVariableColorStyle("q", ["context", "q"])).toEqual(getVariableColorStyle("quest", ["context", "quest"]));
  });
});
