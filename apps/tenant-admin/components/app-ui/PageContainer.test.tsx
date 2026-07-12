import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PageContainer } from "./PageContainer";

describe("PageContainer", () => {
  it("uses the standard wide page layout by default", () => {
    render(<PageContainer aria-label="Settings page">Content</PageContainer>);
    expect(screen.getByLabelText("Settings page")).toHaveClass("max-w-7xl", "gap-6", "px-4", "py-6", "md:px-6");
  });

  it("supports readable and relaxed semantic page layouts", () => {
    render(<PageContainer as="main" size="readable" spacing="relaxed" aria-label="Readable page">Content</PageContainer>);
    expect(screen.getByRole("main", { name: "Readable page" })).toHaveClass("max-w-5xl", "gap-8");
  });

  it("supports embedded full-width layouts without duplicate padding", () => {
    render(<PageContainer size="full" padding="none" aria-label="Embedded content">Content</PageContainer>);
    const container = screen.getByLabelText("Embedded content");
    expect(container).toHaveClass("max-w-none", "p-0");
    expect(container).not.toHaveClass("px-4", "py-6");
  });
});
