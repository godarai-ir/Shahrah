import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DebouncedTextFilter, RangeFilterField } from "./FilterField";

describe("filter fields", () => {
  it("debounces text filter changes", () => {
    vi.useFakeTimers();
    const onValueChange = vi.fn();
    render(<DebouncedTextFilter label="Request ID" value="" onValueChange={onValueChange} />);
    fireEvent.change(screen.getByLabelText("Request ID"), { target: { value: "req-1" } });
    expect(onValueChange).not.toHaveBeenCalled();
    act(() => vi.advanceTimersByTime(350));
    expect(onValueChange).toHaveBeenCalledWith("req-1");
    vi.useRealTimers();
  });

  it("cancels a pending draft when the controlled value is reset", () => {
    vi.useFakeTimers();
    const onValueChange = vi.fn();
    const { rerender } = render(
      <DebouncedTextFilter label="Request ID" value="req-1" onValueChange={onValueChange} />,
    );

    fireEvent.change(screen.getByLabelText("Request ID"), { target: { value: "stale" } });
    rerender(<DebouncedTextFilter label="Request ID" value="" onValueChange={onValueChange} />);
    expect(screen.getByLabelText("Request ID")).toHaveValue("");

    act(() => vi.advanceTimersByTime(350));
    expect(onValueChange).not.toHaveBeenCalled();
    vi.useRealTimers();
  });

  it("does not restore a committed value after an external clear", () => {
    vi.useFakeTimers();
    const onValueChange = vi.fn();
    const { rerender } = render(
      <DebouncedTextFilter label="Request ID" value="" onValueChange={onValueChange} />,
    );

    fireEvent.change(screen.getByLabelText("Request ID"), { target: { value: "req-1" } });
    act(() => vi.advanceTimersByTime(350));
    rerender(<DebouncedTextFilter label="Request ID" value="req-1" onValueChange={onValueChange} />);
    onValueChange.mockClear();
    rerender(<DebouncedTextFilter label="Request ID" value="" onValueChange={onValueChange} />);

    act(() => vi.advanceTimersByTime(350));
    expect(screen.getByLabelText("Request ID")).toHaveValue("");
    expect(onValueChange).not.toHaveBeenCalled();
    vi.useRealTimers();
  });

  it("labels a controlled range slider", () => {
    render(<RangeFilterField label="Latency" value={[10, 90]} min={0} max={100} minLabel="Minimum" maxLabel="Maximum" formatValue={String} onValueCommit={() => undefined} />);
    expect(screen.getAllByRole("slider")).toHaveLength(2);
  });
});
