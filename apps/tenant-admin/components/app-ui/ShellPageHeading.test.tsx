import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ShellPageHeading } from "./ShellPageHeading";

describe("ShellPageHeading", () => {
  it("renders the single route-level heading and description", () => {
    render(<ShellPageHeading title="Rules" description="Manage gateway rules." />);

    expect(screen.getByRole("heading", { level: 1, name: "Rules" })).toBeInTheDocument();
    expect(screen.getByText("Manage gateway rules.")).toHaveAttribute("title", "Manage gateway rules.");
  });

  it("supports title-only pages", () => {
    render(<ShellPageHeading title="Plan" />);

    expect(screen.getByRole("heading", { level: 1, name: "Plan" })).toBeInTheDocument();
    expect(screen.queryByRole("paragraph")).not.toBeInTheDocument();
  });

  it("can render a non-heading preview without changing the visual contract", () => {
    render(<ShellPageHeading as="div" title="Preview" description="Design-system example." />);

    expect(screen.queryByRole("heading")).not.toBeInTheDocument();
    expect(screen.getByText("Preview")).toBeInTheDocument();
  });
});
