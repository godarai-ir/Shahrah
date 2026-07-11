import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AppSelectField } from "./AppSelectField";

describe("AppSelectField", () => {
  it("connects its label, helper text, and error to the select trigger", () => {
    render(
      <AppSelectField
        id="status"
        label="Status"
        description="Choose a status."
        error="Status is required."
        value=""
        onValueChange={vi.fn()}
        options={[{ value: "active", label: "Active" }]}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: "Status" });
    expect(trigger).toHaveAttribute("aria-invalid", "true");
    expect(trigger.getAttribute("aria-describedby")).toContain("status-description");
    expect(trigger.getAttribute("aria-describedby")).toContain("status-error");
    expect(screen.getByRole("alert")).toHaveTextContent("Status is required.");
  });
});
