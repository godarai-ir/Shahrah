import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { FormSheet } from "./FormSheet";

describe("FormSheet", () => {
  it("renders accessible form sheet structure", () => {
    render(<FormSheet open onOpenChange={vi.fn()} title="Create team" description="Team details"><form>Form body</form></FormSheet>);
    expect(screen.getByRole("dialog", { name: "Create team" })).toHaveTextContent("Team details");
    expect(screen.getByText("Form body")).toBeInTheDocument();
  });

  it("prevents dismissal while pending", () => {
    const onOpenChange = vi.fn();
    render(<FormSheet open pending onOpenChange={onOpenChange} title="Saving" description="Please wait">Body</FormSheet>);
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onOpenChange).not.toHaveBeenCalled();
    expect(screen.getByRole("dialog")).toHaveAttribute("aria-busy", "true");
  });
});
