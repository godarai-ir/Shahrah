import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { formatJson, JsonEditor } from "./JsonEditor";

describe("JsonEditor", () => {
  it("renders editable JSON with a direction-safe highlighted overlay", () => {
    const { container } = render(
      <JsonEditor aria-label="Request JSON" value={'{"message":"سلام {{context}}"}'} variables={["context"]} onValueChange={() => undefined} />,
    );

    const textarea = screen.getByRole("textbox", { name: "Request JSON" });
    expect(textarea).toHaveAttribute("dir", "ltr");
    expect(textarea).toHaveClass("peer", "text-transparent", "caret-foreground");
    expect(textarea.nextElementSibling).toHaveClass("peer-focus-visible:ring-inset");
    expect(container.querySelector(".variable-placeholder")).toHaveTextContent("{{context}}");
  });

  it("formats valid JSON on blur and leaves invalid drafts unchanged", () => {
    const onValueChange = vi.fn();
    const value =
      '{"messages":[{"role":"system","content":"You are a precise assistant that returns a detailed structured response."},{"role":"user","content":"Write a complete introduction to Tehran for a first-time visitor."}],"stream":false}';
    const { rerender } = render(<JsonEditor aria-label="Request JSON" value={value} onValueChange={onValueChange} />);

    fireEvent.blur(screen.getByRole("textbox", { name: "Request JSON" }));
    expect(onValueChange).toHaveBeenCalledWith(formatJson(value));
    expect(formatJson(value)).toContain("\n");

    onValueChange.mockClear();
    rerender(<JsonEditor aria-label="Request JSON" value="{" onValueChange={onValueChange} />);
    fireEvent.blur(screen.getByRole("textbox", { name: "Request JSON" }));
    expect(onValueChange).not.toHaveBeenCalled();
  });
});
