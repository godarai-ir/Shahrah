import { fireEvent, render, screen } from "@testing-library/react";
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

  it("renders labelled option groups and disabled options", () => {
    render(
      <AppSelectField
        label="Provider"
        value="openai"
        onValueChange={vi.fn()}
        groups={[
          {
            label: "Available",
            options: [{ value: "openai", label: "OpenAI" }],
          },
          {
            label: "Unavailable",
            options: [{ value: "anthropic", label: "Anthropic", disabled: true }],
          },
        ]}
      />,
    );

    fireEvent.click(screen.getByRole("combobox", { name: "Provider" }));

    expect(screen.getByText("Available")).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Anthropic" })).toHaveAttribute("data-disabled");
  });

  it("supports the compact control size", () => {
    render(
      <AppSelectField
        ariaLabel="Page size"
        value="10"
        onValueChange={vi.fn()}
        size="sm"
        options={[{ value: "10", label: "10" }]}
      />,
    );

    expect(screen.getByRole("combobox", { name: "Page size" })).toHaveClass("h-8", "text-xs");
  });
});
