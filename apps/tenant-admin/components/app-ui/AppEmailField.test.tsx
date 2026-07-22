import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AppEmailField, isValidEmail } from "./AppEmailField";

describe("AppEmailField", () => {
  it("validates common email addresses", () => {
    expect(isValidEmail("user@example.com")).toBe(true);
    expect(isValidEmail(" user@example.com ")).toBe(true);
    expect(isValidEmail("user@example")).toBe(false);
    expect(isValidEmail("user example.com")).toBe(false);
  });

  it("always renders an LTR email input", () => {
    render(<AppEmailField label="ایمیل" invalidMessage="ایمیل معتبر نیست." />);

    const input = screen.getByRole("textbox", { name: "ایمیل" });
    expect(input).toHaveAttribute("type", "email");
    expect(input).toHaveAttribute("dir", "ltr");
    expect(input).toHaveAttribute("inputmode", "email");
    expect(input).toHaveAttribute("autocomplete", "email");
  });

  it("shows accessible client-side validation after blur", () => {
    render(<AppEmailField label="ایمیل" invalidMessage="ایمیل معتبر نیست." requiredMessage="ایمیل الزامی است." required />);

    const input = screen.getByRole("textbox", { name: "ایمیل" });
    fireEvent.change(input, { target: { value: "invalid-email" } });
    fireEvent.blur(input);

    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("alert")).toHaveTextContent("ایمیل معتبر نیست.");
  });
});
