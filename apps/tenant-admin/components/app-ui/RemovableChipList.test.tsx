import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { RemovableChipList } from "./RemovableChipList";

describe("RemovableChipList", () => {
  it("removes a controlled chip by value", () => {
    const onRemove = vi.fn();
    render(<RemovableChipList items={[{ value: "500", label: "Server error", removeLabel: "Remove server error" }]} onRemove={onRemove} />);
    expect(screen.getByText("Server error")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Remove server error" }));
    expect(onRemove).toHaveBeenCalledWith("500");
  });
});
