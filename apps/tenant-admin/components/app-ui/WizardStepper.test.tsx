import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { WizardStepper } from "./WizardStepper";

vi.mock("next-intl", async (importOriginal) => ({
  ...(await importOriginal<typeof import("next-intl")>()),
  useLocale: () => "en",
}));

describe("WizardStepper", () => {
  it("marks the active step and only selects reachable steps", () => {
    const onPast = vi.fn();
    const onActive = vi.fn();
    const onFuture = vi.fn();
    render(
      <WizardStepper
        aria-label="Steps"
        steps={[
          { id: "type", title: "Type", description: "Regex", state: "complete", onSelect: onPast },
          { id: "config", title: "Settings", state: "active", onSelect: onActive },
          { id: "review", title: "Review", state: "idle", disabled: true, onSelect: onFuture },
        ]}
      />,
    );

    expect(screen.getByRole("list", { name: "Steps" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Settings/ })).toHaveAttribute("aria-current", "step");
    expect(screen.getByRole("button", { name: /Review/ })).toBeDisabled();
    expect(screen.getByRole("button", { name: /Type/ })).toHaveTextContent("Regex");

    fireEvent.click(screen.getByRole("button", { name: /Type/ }));
    fireEvent.click(screen.getByRole("button", { name: /Settings/ }));
    fireEvent.click(screen.getByRole("button", { name: /Review/ }));

    expect(onPast).toHaveBeenCalledTimes(1);
    expect(onActive).not.toHaveBeenCalled();
    expect(onFuture).not.toHaveBeenCalled();
  });
});
