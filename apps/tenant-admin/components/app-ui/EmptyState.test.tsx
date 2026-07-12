import { render, screen } from "@testing-library/react";
import { Inbox } from "lucide-react";
import { describe, expect, it } from "vitest";

import { EmptyState } from "./EmptyState";

describe("EmptyState", () => {
  it("uses the surface variant by default", () => {
    render(<EmptyState icon={Inbox} title="No records" description="Create the first record." />);

    const state = screen.getByRole("heading", { name: "No records" }).parentElement?.parentElement;
    expect(state).toHaveAttribute("data-variant", "surface");
    expect(state).toHaveClass("border-dashed", "bg-card", "min-h-48");
    expect(screen.getByText("Create the first record.")).toBeInTheDocument();
  });

  it.each([
    ["inline", "min-h-32"],
    ["table", "min-h-40"],
  ] as const)("renders the %s variant without a nested surface", (variant, minHeightClass) => {
    render(<EmptyState variant={variant} density="compact" title={`${variant} empty`} />);

    const state = screen.getByRole("heading", { name: `${variant} empty` }).parentElement?.parentElement;
    expect(state).toHaveAttribute("data-variant", variant);
    expect(state).toHaveAttribute("data-density", "compact");
    expect(state).toHaveClass(minHeightClass);
    expect(state).not.toHaveClass("border", "bg-card");
  });

  it("renders primary and secondary actions", () => {
    render(
      <EmptyState
        title="No records"
        primaryAction={<button type="button">Create</button>}
        secondaryAction={<button type="button">Learn more</button>}
      />,
    );

    expect(screen.getByRole("button", { name: "Create" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Learn more" })).toBeInTheDocument();
  });
});
