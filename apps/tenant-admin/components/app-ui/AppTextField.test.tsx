import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AppTextField, AppTextareaField } from "./AppTextField";

describe("AppTextField", () => {
  it("connects labels, descriptions, and errors", () => {
    render(<AppTextField label="Name" description="Public name" error="Name is required" />);
    const input = screen.getByRole("textbox", { name: "Name" });
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByText("Public name")).toBeInTheDocument();
    expect(screen.getByText("Name is required")).toBeInTheDocument();
  });

  it("supports textarea controls and disabled semantics", () => {
    render(<AppTextareaField label="Description" disabled />);
    expect(screen.getByRole("textbox", { name: "Description" })).toBeDisabled();
  });
});
