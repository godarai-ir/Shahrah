import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DebouncedTextFilter, RangeFilterField } from "./FilterField";

describe("filter fields", () => {
  it("debounces text filter changes", () => {
    vi.useFakeTimers();
    const onValueChange = vi.fn();
    render(<DebouncedTextFilter label="Request ID" value="" onValueChange={onValueChange} />);
    fireEvent.change(screen.getByLabelText("Request ID"), { target: { value: "req-1" } });
    expect(onValueChange).not.toHaveBeenCalled();
    vi.advanceTimersByTime(350);
    expect(onValueChange).toHaveBeenCalledWith("req-1");
    vi.useRealTimers();
  });

  it("labels a controlled range slider", () => {
    render(<RangeFilterField label="Latency" value={[10, 90]} min={0} max={100} minLabel="Minimum" maxLabel="Maximum" formatValue={String} onValueCommit={() => undefined} />);
    expect(screen.getAllByRole("slider")).toHaveLength(2);
  });
});
