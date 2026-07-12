import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Button } from "./button";
import { Spinner } from "./spinner";

describe("Spinner", () => {
  it("exposes a localized status label when used standalone", () => {
    render(<Spinner label="Loading records" size="lg" />);
    expect(screen.getByRole("status", { name: "Loading records" })).toHaveClass("animate-spin", "size-6");
  });

  it("is decorative when the parent already communicates pending state", () => {
    const { container } = render(
      <Button disabled aria-busy="true">
        <Spinner data-icon="inline-start" />
        Saving changes
      </Button>,
    );

    expect(screen.getByRole("button", { name: "Saving changes" })).toHaveAttribute("aria-busy", "true");
    expect(screen.getByText("Saving changes")).toBeVisible();
    expect(container.querySelector('[data-slot="spinner"]')).toHaveAttribute("aria-hidden", "true");
  });
});
