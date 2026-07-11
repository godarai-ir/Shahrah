import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DataListPanel } from "./DataListPanel";
import { TablePagination } from "./TablePagination";

describe("DataListPanel", () => {
  it("switches between loading, empty, and content states", () => {
    const { rerender } = render(
      <DataListPanel loading loadingLabel="Loading resources" emptyState={<p>No resources</p>}>
        <p>Resource table</p>
      </DataListPanel>,
    );

    expect(screen.getByRole("status", { name: "Loading resources" })).toBeInTheDocument();
    expect(screen.queryByText("Resource table")).not.toBeInTheDocument();

    rerender(
      <DataListPanel empty emptyState={<p>No resources</p>}>
        <p>Resource table</p>
      </DataListPanel>,
    );
    expect(screen.getByText("No resources")).toBeInTheDocument();

    rerender(
      <DataListPanel>
        <p>Resource table</p>
      </DataListPanel>,
    );
    expect(screen.getByText("Resource table")).toBeInTheDocument();
  });

  it("renders toolbar metadata and pagination slots", () => {
    render(
      <DataListPanel toolbar={<input aria-label="Search" />} toolbarMeta="2 results" pagination={<p>Pagination</p>}>
        <p>Rows</p>
      </DataListPanel>,
    );

    expect(screen.getByRole("textbox", { name: "Search" })).toBeInTheDocument();
    expect(screen.getByText("2 results")).toBeInTheDocument();
    expect(screen.getByText("Pagination")).toBeInTheDocument();
  });
});

describe("TablePagination", () => {
  it("exposes navigation state and invokes enabled actions", () => {
    const onPrevious = vi.fn();
    const onNext = vi.fn();

    render(
      <TablePagination
        label="Resource pages"
        summary="1–25 of 50"
        page={1}
        pageSize={25}
        pageSizeLabel="Rows"
        previousLabel="Previous"
        nextLabel="Next"
        canPrevious={false}
        canNext
        onPrevious={onPrevious}
        onNext={onNext}
      />,
    );

    expect(screen.getByRole("navigation", { name: "Resource pages" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(onPrevious).not.toHaveBeenCalled();
    expect(onNext).toHaveBeenCalledOnce();
  });
});
