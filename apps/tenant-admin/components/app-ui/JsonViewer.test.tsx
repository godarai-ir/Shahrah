import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { JsonViewer, serializeJson } from "./JsonViewer";

describe("JsonViewer", () => {
  it("renders formatted, read-only JSON with CodeMirror navigation affordances", () => {
    const { container } = render(
      <JsonViewer aria-label="Log summary JSON" value={{ log_id: "log_123", cached: false }} />,
    );

    const viewer = screen.getByRole("textbox", { name: "Log summary JSON" });
    expect(viewer).toHaveAttribute("contenteditable", "false");
    expect(viewer).toHaveAttribute("aria-readonly", "true");
    expect(viewer).toHaveAttribute("dir", "ltr");
    expect(container.querySelector('[data-slot="json-viewer"] .cm-editor')).toBeInTheDocument();
    expect(container.querySelector(".cm-lineNumbers")).toBeInTheDocument();
    expect(container.querySelector(".cm-foldGutter")).toBeInTheDocument();
    expect(container).toHaveTextContent('"log_id"');
    expect(container).toHaveTextContent('"log_123"');
  });

  it("falls back safely for values that cannot be stringified", () => {
    const circular: Record<string, unknown> = {};
    circular.self = circular;

    expect(serializeJson(circular)).toBe("[object Object]");
  });
});
