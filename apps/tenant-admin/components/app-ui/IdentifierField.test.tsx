import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

import { IdentifierField, isValidIdentifier } from "./IdentifierField";

describe("IdentifierField", () => {
  it("accepts only the shared identifier character set", () => {
    expect(isValidIdentifier("prompt_name-2")).toBe(true);
    expect(isValidIdentifier("prompt name")).toBe(false);
    expect(isValidIdentifier("پرامپت")).toBe(false);
    expect(isValidIdentifier("")).toBe(false);
  });

  it("renders an accessible LTR input with helper text", () => {
    const onChange = vi.fn();
    render(<IdentifierField label="شناسه" value="team_one" onChange={onChange} required />);
    const input = screen.getByLabelText(/شناسه/);
    expect(input).toHaveAttribute("dir", "ltr");
    expect(input).toHaveAttribute("pattern", "[A-Za-z0-9_-]+");
    expect(input).toHaveAttribute("aria-describedby");
    expect(screen.getByText("identifierField.description")).toBeInTheDocument();
    fireEvent.change(input, { target: { value: "team_two" } });
    expect(onChange).toHaveBeenCalled();
  });

  it("connects an error to the input", () => {
    render(<IdentifierField label="شناسه" error="شناسه نامعتبر است." />);
    const input = screen.getByLabelText("شناسه");
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input.getAttribute("aria-describedby")).toContain("-error");
    expect(screen.getByRole("alert")).toHaveTextContent("شناسه نامعتبر است.");
  });
});
