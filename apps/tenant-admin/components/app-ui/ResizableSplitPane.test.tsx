import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { ResizableSplitPane } from "./ResizableSplitPane";

afterEach(cleanup);

describe("ResizableSplitPane", () => {
  it("exposes an accessible separator and supports RTL arrow keys", () => {
    render(<ResizableSplitPane direction="rtl" start={<div>Prompt</div>} end={<div>Response</div>} startLabel="Prompt" endLabel="Response" separatorLabel="Resize panes" />);
    const separator = screen.getByRole("separator", { name: "Resize panes" });
    expect(separator).toHaveAttribute("aria-valuenow", "50");
    fireEvent.keyDown(separator, { key: "ArrowLeft" });
    expect(separator).toHaveAttribute("aria-valuenow", "52");
  });

  it("restores existing body interaction styles when a drag is cancelled", () => {
    document.body.style.cursor = "wait";
    document.body.style.userSelect = "text";
    render(<ResizableSplitPane start={<div>Prompt</div>} end={<div>Response</div>} startLabel="Prompt" endLabel="Response" separatorLabel="Resize panes" />);
    const separator = screen.getByRole("separator", { name: "Resize panes" });

    fireEvent.pointerDown(separator, { clientX: 100 });
    expect(document.body.style.cursor).toBe("col-resize");
    expect(document.body.style.userSelect).toBe("none");
    fireEvent.pointerCancel(window);

    expect(document.body.style.cursor).toBe("wait");
    expect(document.body.style.userSelect).toBe("text");
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  });
});
