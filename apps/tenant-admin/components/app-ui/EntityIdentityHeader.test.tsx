import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { EntityIdentityHeader } from "./EntityIdentityHeader";

describe("EntityIdentityHeader", () => {
  it("renders the identity, supporting content, and actions", () => {
    render(<EntityIdentityHeader title="Production user" identifier="vu_123" description="Service identity" status={<span>Active</span>} actions={<button>Edit</button>} copy={{ copyLabel: "Copy ID", copiedLabel: "Copied" }} />);
    expect(screen.getByText("Production user")).toBeInTheDocument();
    expect(screen.getByText("vu_123")).toHaveAttribute("dir", "ltr");
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Copy ID" })).toBeInTheDocument();
  });
});
