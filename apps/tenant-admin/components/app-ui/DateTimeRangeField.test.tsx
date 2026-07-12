import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DateTimeRangeField } from "./DateTimeRangeField";

describe("DateTimeRangeField", () => {
  it("renders labelled LTR controls and emits ISO values", () => {
    const onFromChange = vi.fn();
    render(<DateTimeRangeField from="2025-01-02T03:04:00.000Z" to="2025-01-03T03:04:00.000Z" fromLabel="From" toLabel="To" onFromChange={onFromChange} onToChange={vi.fn()} />);

    const from = screen.getByLabelText("From");
    expect(from).toHaveAttribute("type", "datetime-local");
    expect(from).toHaveAttribute("dir", "ltr");
    fireEvent.change(from, { target: { value: "2025-02-03T10:30" } });
    expect(onFromChange).toHaveBeenCalledWith(expect.stringMatching(/^2025-02-03T/));
  });
});
