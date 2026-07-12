import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ActiveFilterChips } from "./ActiveFilterChips";

describe("ActiveFilterChips", () => {
  it("renders removable filter values", () => {
    const onRemove = vi.fn();
    render(<ActiveFilterChips items={[{ key: "model", label: "Model", value: "gpt" }]} onRemove={onRemove} getRemoveLabel={(item) => `Remove ${item.label}`} />);
    fireEvent.click(screen.getByRole("button", { name: "Remove Model" }));
    expect(onRemove).toHaveBeenCalledWith("model");
  });
});
