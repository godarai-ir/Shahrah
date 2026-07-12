import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { formatDateTime } from "@/lib/date-time";
import { DateTimeValue } from "./DateTimeValue";

describe("DateTimeValue", () => {
  it("renders a machine-readable time element", () => {
    render(<DateTimeValue value="2025-01-02T03:04:00.000Z" locale="en-US" format={{ dateStyle: "medium" }} />);
    expect(screen.getByText(/Jan/)).toHaveAttribute("datetime", "2025-01-02T03:04:00.000Z");
  });

  it("uses the requested fallback for invalid values", () => {
    render(<DateTimeValue value="invalid" fallback="Not set" />);
    expect(screen.getByText("Not set")).not.toHaveAttribute("datetime");
    expect(formatDateTime(undefined, { fallback: "None" })).toBe("None");
  });
});
