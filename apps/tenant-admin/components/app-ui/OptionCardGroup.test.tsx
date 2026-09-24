import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { OptionCardGroup } from "./OptionCardGroup";

const options = [
  { value: "enforce", title: "Block and log", description: "The request is stopped." },
  { value: "audit", title: "Allow and flag", description: "The request passes.", tip: "Good for trials." },
  { value: "off", title: "Off", description: "Never runs.", disabled: true },
] as const;

describe("OptionCardGroup", () => {
  it("renders a labelled radiogroup of described radio cards", () => {
    render(<OptionCardGroup aria-label="Enforcement" options={options} value="enforce" onValueChange={() => undefined} />);

    expect(screen.getByRole("radiogroup", { name: "Enforcement" })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Block and log" })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByRole("radio", { name: "Allow and flag" })).toHaveAttribute("aria-checked", "false");
    expect(screen.getByRole("radio", { name: "Allow and flag" })).toHaveAccessibleDescription("The request passes.Good for trials.");
  });

  it("selects a card on click and ignores disabled cards", () => {
    const onValueChange = vi.fn();
    render(<OptionCardGroup aria-label="Enforcement" options={options} value="enforce" onValueChange={onValueChange} />);

    fireEvent.click(screen.getByRole("radio", { name: "Allow and flag" }));
    fireEvent.click(screen.getByRole("radio", { name: "Off" }));

    expect(onValueChange).toHaveBeenCalledTimes(1);
    expect(onValueChange).toHaveBeenCalledWith("audit");
    expect(screen.getByRole("radio", { name: "Off" })).toBeDisabled();
  });
});
