import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { OneTimeSecretDialog } from "./OneTimeSecretDialog";

describe("OneTimeSecretDialog", () => {
  it("clears the secret whenever the dialog closes", () => {
    const onOpenChange = vi.fn();
    const onClear = vi.fn();
    render(
      <OneTimeSecretDialog
        open
        onOpenChange={onOpenChange}
        value="one-time-secret"
        title="Copy secret"
        description="This value is shown once."
        warning="Store it securely."
        closeLabel="Close"
        copyLabels={{
          valueAriaLabel: "Secret value",
          copy: "Copy",
          copying: "Copying",
          copied: "Copied",
        }}
        onClear={onClear}
      />,
    );

    expect(screen.getByRole("dialog", { name: "Copy secret" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("one-time-secret")).toHaveAttribute("dir", "ltr");
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onClear).toHaveBeenCalledOnce();
  });
});
