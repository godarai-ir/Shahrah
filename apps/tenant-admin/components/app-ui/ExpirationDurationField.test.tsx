import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ExpirationDurationField } from "./ExpirationDurationField";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("ExpirationDurationField", () => {
  it("renders an accessible LTR duration field with several examples", () => {
    render(<ExpirationDurationField id="expiry" />);

    const input = screen.getByRole("textbox", { name: "expirationDuration.label" });
    expect(input).toHaveAttribute("dir", "ltr");
    expect(input).toHaveAttribute("aria-describedby", "expiry-examples");
    expect(screen.getByText("30m")).toBeInTheDocument();
    expect(screen.getByText("24h")).toBeInTheDocument();
    expect(screen.getByText("7d")).toBeInTheDocument();
    expect(screen.getByText("30d")).toBeInTheDocument();
  });
});
