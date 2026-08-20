import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DocumentIngestionStatus, processingDocumentStatuses } from "./DocumentIngestionStatus";

describe("DocumentIngestionStatus", () => {
  it("shows the current processing stage and progress", () => {
    render(<DocumentIngestionStatus status="embedding" label="بردارسازی" />);

    expect(screen.getByText("بردارسازی")).toBeInTheDocument();
    expect(screen.getByRole("progressbar", { name: "بردارسازی" })).toHaveAttribute("aria-valuenow", "70");
    expect(processingDocumentStatuses.has("embedding")).toBe(true);
  });

  it("marks a newly transitioned status for its micro-interaction", () => {
    const { container } = render(<DocumentIngestionStatus status="ready" label="آماده" changed />);

    expect(container.querySelector("[data-status='ready']")).toHaveAttribute("data-status-changed", "true");
    expect(screen.getByRole("progressbar", { name: "آماده" })).toHaveAttribute("aria-valuenow", "100");
    expect(processingDocumentStatuses.has("ready")).toBe(false);
  });
});
