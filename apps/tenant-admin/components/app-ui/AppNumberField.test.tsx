import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AppNumberField } from "./AppNumberField";

describe("AppNumberField", () => {
  it("applies numeric and LTR defaults", () => {
    render(<AppNumberField label="Timeout" min={1} step={1} />);
    const input = screen.getByRole("spinbutton", { name: "Timeout" });
    expect(input).toHaveAttribute("dir", "ltr");
    expect(input).toHaveAttribute("inputmode", "decimal");
  });
});
