import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RefreshButton } from "./RefreshButton";

describe("RefreshButton", () => {
  it("renders an icon-only accessible refresh action", () => {
    render(<RefreshButton label="Refresh" />);

    const button = screen.getByRole("button", { name: "Refresh" });
    expect(button).toBeInTheDocument();
    expect(button).not.toHaveTextContent("Refresh");
  });

  it("announces and disables its loading state", () => {
    render(<RefreshButton label="Refresh" loadingLabel="Refreshing" isLoading />);

    const button = screen.getByRole("button", { name: "Refreshing" });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("aria-busy", "true");
  });
});
