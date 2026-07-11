import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { CommercialPlanUpgradeDialog } from "./CommercialPlanUpgradeDialog";

describe("CommercialPlanUpgradeDialog", () => {
  afterEach(cleanup);

  it("renders consistent plan guidance and closes from both actions", () => {
    const onOpenChange = vi.fn();

    render(
      <CommercialPlanUpgradeDialog
        open
        onOpenChange={onOpenChange}
        title="Feature unavailable"
        description="Upgrade your plan to continue."
        cancelLabel="Not now"
        actionLabel="Upgrade plan"
      />,
    );

    expect(screen.getByRole("dialog")).toHaveTextContent("Upgrade your plan to continue.");
    expect(screen.getByRole("link", { name: "Upgrade plan" })).toHaveAttribute("href", "/billing/plan");

    fireEvent.click(screen.getByRole("button", { name: "Not now" }));
    fireEvent.click(screen.getByRole("link", { name: "Upgrade plan" }));

    expect(onOpenChange).toHaveBeenCalledTimes(2);
    expect(onOpenChange).toHaveBeenNthCalledWith(1, false);
    expect(onOpenChange).toHaveBeenNthCalledWith(2, false);
  });
});
