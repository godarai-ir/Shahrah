import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";

import { ConfirmDialog } from "./ConfirmDialog";

describe("ConfirmDialog", () => {
  it("supports trigger-based confirmation and closes after success", async () => {
    const onConfirm = vi.fn();

    render(
      <ConfirmDialog
        trigger={<Button>Open confirmation</Button>}
        title="Delete resource"
        description="This action cannot be undone."
        cancelLabel="Cancel"
        confirmLabel="Delete"
        onConfirm={onConfirm}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open confirmation" }));
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(onConfirm).toHaveBeenCalledOnce();
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("shows pending state and prevents dismissal until async confirmation succeeds", async () => {
    let resolveConfirmation!: () => void;
    const confirmation = new Promise<void>((resolve) => {
      resolveConfirmation = resolve;
    });

    render(
      <ConfirmDialog
        trigger={<Button>Open confirmation</Button>}
        title="Revoke token"
        cancelLabel="Cancel"
        confirmLabel="Revoke"
        pendingLabel="Revoking"
        onConfirm={() => confirmation}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open confirmation" }));
    fireEvent.click(screen.getByRole("button", { name: "Revoke" }));

    expect(screen.getByRole("button", { name: "Revoking" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Close" })).not.toBeInTheDocument();

    await act(async () => resolveConfirmation());
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("keeps the dialog open and reports rejected confirmations", async () => {
    const onConfirmError = vi.fn();

    render(
      <ConfirmDialog
        trigger={<Button>Open confirmation</Button>}
        title="Delete resource"
        cancelLabel="Cancel"
        confirmLabel="Delete"
        onConfirm={() => Promise.reject(new Error("Delete failed"))}
        onConfirmError={onConfirmError}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open confirmation" }));
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(onConfirmError).toHaveBeenCalledOnce());
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete" })).toBeEnabled();
  });
});
