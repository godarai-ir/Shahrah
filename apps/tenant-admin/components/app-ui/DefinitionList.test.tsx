import { render, screen } from "@testing-library/react";
import { Database } from "lucide-react";
import { describe, expect, it } from "vitest";

import { DefinitionList, DefinitionListItem } from "./DefinitionList";

describe("DefinitionList", () => {
  it("renders semantic terms and definitions", () => {
    const { container } = render(
      <DefinitionList columns={2} aria-label="Request metadata">
        <DefinitionListItem label="Provider" value="OpenAI" icon={Database} />
        <DefinitionListItem label="Model" value="gpt-5.4" mono valueDir="ltr" />
      </DefinitionList>,
    );

    expect(screen.getByText("Provider", { selector: "dt" })).toBeInTheDocument();
    expect(screen.getByRole("definition", { name: "OpenAI" })).toHaveAttribute("title", "OpenAI");
    expect(screen.getByRole("definition", { name: "gpt-5.4" })).toHaveAttribute("dir", "ltr");
    expect(container.querySelector("dl")).toHaveClass("sm:grid-cols-2");
  });

  it("supports surfaced, emphasized, and wrapped metadata", () => {
    render(
      <DefinitionList>
        <DefinitionListItem
          label="Request ID"
          value="req_123456789"
          mono
          wrap
          emphasis
          surface
          valueDir="ltr"
        />
      </DefinitionList>,
    );

    const definition = screen.getByText("req_123456789", { selector: "dd" });
    expect(definition).toHaveClass("break-all", "font-mono", "font-semibold");
    expect(definition).not.toHaveAttribute("title");
    expect(definition.parentElement?.parentElement).toHaveClass("border-border");
  });
});
