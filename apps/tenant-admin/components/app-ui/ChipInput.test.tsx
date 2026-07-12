import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ChipInput } from "./ChipInput";

describe("ChipInput", () => {
  it("adds unique parsed values and removes chips", () => {
    const onValueChange = vi.fn();
    render(<ChipInput value={["429"]} onValueChange={onValueChange} getRemoveLabel={(value) => `Remove ${value}`} placeholder="Status" />);
    fireEvent.change(screen.getByPlaceholderText("Status"), { target: { value: "500, 429" } });
    fireEvent.keyDown(screen.getByPlaceholderText("Status"), { key: "Enter" });
    expect(onValueChange).toHaveBeenCalledWith(["429", "500"]);
    fireEvent.click(screen.getByRole("button", { name: "Remove 429" }));
    expect(onValueChange).toHaveBeenLastCalledWith([]);
  });
});
