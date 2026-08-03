import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { UnsavedChangesDialog } from "./UnsavedChangesDialog";

describe("UnsavedChangesDialog", () => {
  it("lets the user stay or discard through an accessible dialog", async () => {
    const onDiscard = vi.fn();
    const onOpenChange = vi.fn();
    render(<UnsavedChangesDialog open onOpenChange={onOpenChange} title="Unsaved changes" description="Your edits will be lost." stayLabel="Keep editing" discardLabel="Discard" onDiscard={onDiscard} />);

    expect(screen.getByRole("alertdialog", { name: "Unsaved changes" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Discard" }));
    expect(onDiscard).toHaveBeenCalledOnce();
  });
});
