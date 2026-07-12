import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AppSwitchField } from "./AppSwitchField";

describe("AppSwitchField", () => {
  it("connects its label, description, and error to the switch", () => {
    const onCheckedChange = vi.fn();
    render(<AppSwitchField label="Enable cache" description="Cache matching requests." error="Review this setting." checked={false} onCheckedChange={onCheckedChange} />);

    const control = screen.getByRole("switch", { name: "Enable cache" });
    expect(control).toHaveAttribute("aria-invalid", "true");
    expect(control).toHaveAccessibleDescription("Cache matching requests. Review this setting.");
    fireEvent.click(control);
    expect(onCheckedChange).toHaveBeenCalledWith(true);
  });

  it("supports an unframed disabled field", () => {
    const { container } = render(<AppSwitchField label="Streaming" checked surface={false} disabled />);
    expect(screen.getByRole("switch", { name: "Streaming" })).toBeDisabled();
    expect(container.querySelector('[data-slot="field"]')).not.toHaveClass("bg-muted/30");
  });
});
