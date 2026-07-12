import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { FormDialog } from "./FormDialog";

describe("FormDialog", () => {
  it("renders accessible structure and blocks dismissal while pending", () => {
    const onOpenChange = vi.fn();
    render(<FormDialog open onOpenChange={onOpenChange} title="Configure cache" description="Update cache behavior." pending footer={<button>Save</button>}><input aria-label="TTL" /></FormDialog>);
    expect(screen.getByRole("dialog", { name: "Configure cache" })).toHaveAccessibleDescription("Update cache behavior.");
    expect(screen.getByLabelText("TTL")).toBeInTheDocument();
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(onOpenChange).not.toHaveBeenCalled();
  });
});
