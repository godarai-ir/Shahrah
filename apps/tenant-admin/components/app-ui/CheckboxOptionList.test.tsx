import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CheckboxOptionList } from "./CheckboxOptionList";

describe("CheckboxOptionList", () => {
  it("returns the next selected values", () => {
    const onValueChange = vi.fn();
    render(<CheckboxOptionList label="Models" options={[{ value: "gpt", label: "GPT" }, { value: "claude", label: "Claude" }]} value={["gpt"]} onValueChange={onValueChange} emptyText="No models" />);
    fireEvent.click(screen.getByRole("checkbox", { name: "Claude" }));
    expect(onValueChange).toHaveBeenCalledWith(["gpt", "claude"]);
  });
});
