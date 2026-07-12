import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { RowActionsMenu } from "./RowActionsMenu";

function openMenu() {
  fireEvent.pointerDown(screen.getByRole("button", { name: "Resource actions" }), {
    button: 0,
    ctrlKey: false,
  });
}

describe("RowActionsMenu", () => {
  it("renders internal and external links in separated action groups", () => {
    render(
      <RowActionsMenu
        label="Resource actions"
        actions={[
          { label: "Edit", href: "/resources/1" },
          { label: "Documentation", href: "https://example.com/docs", external: true },
          { label: "Delete", separatorBefore: true, destructive: true },
        ]}
      />,
    );

    openMenu();

    expect(screen.getByRole("menuitem", { name: "Edit" })).toHaveAttribute("href", "/resources/1");
    expect(screen.getByRole("menuitem", { name: "Documentation" })).toHaveAttribute("target", "_blank");
    expect(screen.getByRole("menuitem", { name: "Documentation" })).toHaveAttribute("rel", "noreferrer");
    expect(document.querySelectorAll('[data-slot="dropdown-menu-separator"]')).toHaveLength(1);
  });

  it("disables a pending action and shows its pending label", () => {
    const onSelect = vi.fn();
    render(
      <RowActionsMenu
        label="Resource actions"
        actions={[{ label: "Disconnect", pendingLabel: "Disconnecting", pending: true, onSelect }]}
      />,
    );

    openMenu();

    const action = screen.getByRole("menuitem", { name: "Disconnecting" });
    expect(action).toHaveAttribute("data-disabled");
    fireEvent.click(action);
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("prevents action-menu clicks from bubbling from a clickable row", () => {
    const onRowClick = vi.fn();
    const onSelect = vi.fn();
    render(
      <div onClick={onRowClick}>
        <RowActionsMenu
          label="Resource actions"
          stopPropagation
          actions={[{ label: "Edit", onSelect }]}
        />
      </div>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Resource actions" }));
    expect(onRowClick).not.toHaveBeenCalled();

    openMenu();
    fireEvent.click(screen.getByRole("menuitem", { name: "Edit" }));
    expect(onSelect).toHaveBeenCalledOnce();
    expect(onRowClick).not.toHaveBeenCalled();
  });
});
