import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AppCodeTextareaField } from "./AppCodeTextareaField";

describe("AppCodeTextareaField", () => {
  it("applies technical text defaults while preserving field semantics", () => {
    render(<AppCodeTextareaField label="JSON schema" description="Enter a valid schema." value={'{"type":"object"}'} readOnly />);

    const textarea = screen.getByRole("textbox", { name: "JSON schema" });
    expect(textarea).toHaveAttribute("dir", "ltr");
    expect(textarea).toHaveAttribute("spellcheck", "false");
    expect(textarea).toHaveClass("font-mono", "text-left");
    expect(textarea).toHaveAccessibleDescription("Enter a valid schema.");
  });
});
