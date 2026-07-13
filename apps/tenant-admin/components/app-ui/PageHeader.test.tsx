import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PageHeader, PageHeaderInShellProvider } from "./PageHeader";

describe("PageHeader", () => {
  it("renders the route heading outside the admin shell", () => {
    render(<PageHeader title="Rules" description="Manage rules." />);

    expect(screen.getByRole("heading", { level: 1, name: "Rules" })).toBeInTheDocument();
    expect(screen.getByText("Manage rules.")).toBeInTheDocument();
  });

  it("suppresses duplicated route metadata when the admin shell owns it", () => {
    render(
      <PageHeaderInShellProvider enabled>
        <PageHeader title="Rules" description="Manage rules." />
      </PageHeaderInShellProvider>,
    );

    expect(screen.queryByRole("heading", { name: "Rules" })).not.toBeInTheDocument();
    expect(screen.queryByText("Manage rules.")).not.toBeInTheDocument();
  });

  it("keeps page actions available when the admin shell owns the heading", () => {
    render(
      <PageHeaderInShellProvider enabled>
        <PageHeader title="Rules" actions={<button type="button">Create rule</button>} />
      </PageHeaderInShellProvider>,
    );

    expect(screen.getByRole("button", { name: "Create rule" })).toBeInTheDocument();
  });
});
