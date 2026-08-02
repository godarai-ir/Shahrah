import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { TypedConfirmDialog } from "./TypedConfirmDialog";

function ControlledDialog({ onConfirm, onConfirmError }: { onConfirm: () => void | Promise<void>; onConfirmError?: (error: unknown) => void }) {
  const [open, setOpen] = useState(true);
  return (
    <TypedConfirmDialog
      open={open}
      onOpenChange={setOpen}
      title="Delete team"
      description="This action cannot be undone."
      expectedValue="platform"
      inputLabel="Type platform to confirm"
      cancelLabel="Cancel"
      confirmLabel="Delete"
      pendingLabel="Deleting"
      onConfirm={onConfirm}
      onConfirmError={onConfirmError}
    />
  );
}

describe("TypedConfirmDialog", () => {
  it("requires an exact typed value before confirmation", () => {
    const onConfirm = vi.fn();
    render(<ControlledDialog onConfirm={onConfirm} />);

    const confirm = screen.getByRole("button", { name: "Delete" });
    expect(confirm).toBeDisabled();
    fireEvent.change(screen.getByLabelText("Type platform to confirm"), { target: { value: "Platform" } });
    expect(confirm).toBeDisabled();
    fireEvent.change(screen.getByLabelText("Type platform to confirm"), { target: { value: "platform" } });
    expect(confirm).toBeEnabled();
  });

  it("shows pending state, prevents dismissal, and closes after success", async () => {
    let resolveConfirmation!: () => void;
    const confirmation = new Promise<void>((resolve) => { resolveConfirmation = resolve; });
    render(<ControlledDialog onConfirm={() => confirmation} />);

    fireEvent.change(screen.getByLabelText("Type platform to confirm"), { target: { value: "platform" } });
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(screen.getByRole("button", { name: "Deleting" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Close" })).not.toBeInTheDocument();

    await act(async () => resolveConfirmation());
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("stays open and reports a failed confirmation", async () => {
    const onConfirmError = vi.fn();
    render(<ControlledDialog onConfirm={() => Promise.reject(new Error("Delete failed"))} onConfirmError={onConfirmError} />);

    fireEvent.change(screen.getByLabelText("Type platform to confirm"), { target: { value: "platform" } });
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(onConfirmError).toHaveBeenCalledOnce());
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete" })).toBeEnabled();
  });

  it("renders operation feedback inside the dialog", () => {
    render(
      <TypedConfirmDialog
        open
        onOpenChange={() => undefined}
        title="Delete team"
        description="This action cannot be undone."
        expectedValue="platform"
        inputLabel="Type platform to confirm"
        cancelLabel="Cancel"
        confirmLabel="Delete"
        pendingLabel="Deleting"
        onConfirm={() => undefined}
        feedback={<div role="alert">The team is still referenced.</div>}
      />,
    );

    const dialog = screen.getByRole("dialog");
    expect(dialog).toContainElement(screen.getByRole("alert"));
    expect(screen.getByText("The team is still referenced.")).toBeInTheDocument();
  });
});
