import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { MarkdownRenderer } from "./MarkdownRenderer";

describe("MarkdownRenderer", () => {
  afterEach(cleanup);

  it("renders CommonMark and GFM content as semantic elements", () => {
    render(
      <MarkdownRenderer
        content={'## Result\n\n**Ready** with `inline` code.\n\n- [x] Markdown\n- [ ] Tables\n\n| Feature | State |\n| --- | --- |\n| Code | ready |\n\n[Docs](https://example.com)'}
      />,
    );

    expect(screen.getByRole("heading", { level: 2, name: "Result" })).toBeInTheDocument();
    expect(screen.getByText("Ready").tagName).toBe("STRONG");
    expect(screen.getByText("inline").tagName).toBe("CODE");
    expect(screen.getAllByRole("checkbox")).toHaveLength(2);
    expect(screen.getByRole("columnheader", { name: "Feature" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Docs" })).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("formats fenced TypeScript code with the shared highlighted code block", () => {
    const { container } = render(
      <MarkdownRenderer content={'```typescript\nconst answer: number = 42;\n```'} />,
    );

    const codeBlock = container.querySelector('pre[dir="ltr"]');
    expect(codeBlock).not.toBeNull();
    expect(codeBlock).toHaveTextContent("const answer: number = 42;");
    expect(container.querySelector(".token.keyword")).toHaveTextContent("const");
    expect(container).not.toHaveTextContent("```typescript");
  });

  it("does not interpret raw HTML from model output", () => {
    const { container } = render(<MarkdownRenderer content={'<script>alert("unsafe")</script>'} />);

    expect(container.querySelector("script")).toBeNull();
    expect(container).toHaveTextContent('<script>alert("unsafe")</script>');
  });
});
