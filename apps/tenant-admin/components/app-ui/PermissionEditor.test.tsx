import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PermissionEditor } from "./PermissionEditor";

const permissions = [
  { key: "role:read", group: "roles", description: "Read roles" },
  { key: "role:manage", group: "roles", description: "Manage roles" },
  { key: "user:read", group: "users", description: "Read users" },
];

const labels = {
  selectedTab: "Selected",
  allTab: "All permissions",
  searchPlaceholder: "Search permissions",
  searchAriaLabel: "Search permissions",
  emptyTitle: "No permissions found",
  selectGroup: "Select group",
  clearGroup: "Clear group",
};

describe("PermissionEditor", () => {
  it("groups permissions and returns the next selection", () => {
    const onValueChange = vi.fn();
    render(
      <PermissionEditor
        permissions={permissions}
        value={["role:read"]}
        onValueChange={onValueChange}
        query=""
        onQueryChange={vi.fn()}
        mode="all"
        onModeChange={vi.fn()}
        labels={labels}
        dir="ltr"
      />,
    );

    expect(screen.getByRole("group", { name: /roles/ })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("checkbox", { name: "Manage roles" }));
    expect(onValueChange).toHaveBeenCalledWith(["role:read", "role:manage"]);
  });

  it("supports controlled search and view mode", () => {
    const onQueryChange = vi.fn();
    const onModeChange = vi.fn();
    render(
      <PermissionEditor
        permissions={permissions}
        value={["role:read"]}
        onValueChange={vi.fn()}
        query=""
        onQueryChange={onQueryChange}
        mode="selected"
        onModeChange={onModeChange}
        labels={labels}
        dir="ltr"
      />,
    );

    expect(screen.getByRole("checkbox", { name: "Read roles" })).toBeInTheDocument();
    expect(screen.queryByRole("checkbox", { name: "Manage roles" })).not.toBeInTheDocument();
    fireEvent.change(screen.getByRole("searchbox", { name: "Search permissions" }), { target: { value: "users" } });
    fireEvent.mouseDown(screen.getByRole("tab", { name: "All permissions" }), { button: 0, ctrlKey: false });
    expect(onQueryChange).toHaveBeenCalledWith("users");
    expect(onModeChange).toHaveBeenCalledWith("all");
  });
});
