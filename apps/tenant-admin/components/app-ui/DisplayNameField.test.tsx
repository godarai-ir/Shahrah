import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DisplayNameField, resolveDisplayNameDirection } from "./DisplayNameField";

describe("DisplayNameField", () => {
  it("resolves direction from the first meaningful letter", () => {
    expect(resolveDisplayNameDirection("تیم محصول")).toBe("rtl");
    expect(resolveDisplayNameDirection("עברית")).toBe("rtl");
    expect(resolveDisplayNameDirection("Production team")).toBe("ltr");
    expect(resolveDisplayNameDirection("  42 — تیم محصول")).toBe("rtl");
    expect(resolveDisplayNameDirection("  42 — Production team")).toBe("ltr");
    expect(resolveDisplayNameDirection("42 —")).toBeUndefined();
  });

  it("updates the input direction as the display name changes", () => {
    render(<DisplayNameField label="نام نمایشی" />);

    const input = screen.getByRole("textbox", { name: "نام نمایشی" });
    expect(input).toHaveAttribute("dir", "auto");

    fireEvent.change(input, { target: { value: "Workspace admin" } });
    expect(input).toHaveAttribute("dir", "ltr");

    fireEvent.change(input, { target: { value: "مدیر فضای کاری" } });
    expect(input).toHaveAttribute("dir", "rtl");
  });

  it("uses the direction of a controlled value", () => {
    const { rerender } = render(<DisplayNameField label="Display name" value="English name" readOnly />);
    expect(screen.getByRole("textbox", { name: "Display name" })).toHaveAttribute("dir", "ltr");

    rerender(<DisplayNameField label="Display name" value="نام فارسی" readOnly />);
    expect(screen.getByRole("textbox", { name: "Display name" })).toHaveAttribute("dir", "rtl");
  });
});
