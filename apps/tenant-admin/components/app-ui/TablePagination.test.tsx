import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TablePagination } from "./TablePagination";

const baseProps = {
  label: "Resource pages",
  summary: "Showing 1 to 2 of 2",
  previousLabel: "Previous",
  nextLabel: "Next",
  onPrevious: vi.fn(),
  onNext: vi.fn(),
};

describe("TablePagination", () => {
  it("stays hidden when every row fits on the only page", () => {
    const { container } = render(
      <TablePagination
        {...baseProps}
        canPrevious={false}
        canNext={false}
      />,
    );

    expect(container).toBeEmptyDOMElement();
  });

  it("renders navigation when another page is available", () => {
    render(
      <TablePagination
        {...baseProps}
        canPrevious={false}
        canNext
        page={1}
      />,
    );

    expect(screen.getByText(baseProps.summary)).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: baseProps.label })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: baseProps.previousLabel })).toBeDisabled();
    expect(screen.getByRole("button", { name: baseProps.nextLabel })).toBeEnabled();
  });
});
