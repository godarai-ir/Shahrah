import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PermissionSummary } from "./PermissionSummary";

describe("PermissionSummary", () => {
  it("summarizes selected permissions and clears them", () => {
    const onClear = vi.fn();
    render(
      <PermissionSummary
        permissions={[
          { key: "role:read", group: "roles", description: "Read roles" },
          { key: "user:read", group: "users", description: "Read users" },
        ]}
        value={["role:read", "user:read"]}
        title="Role summary"
        labels={{
          selectedPermissions: "Selected permissions",
          coveredGroups: "Covered groups",
          clearAll: "Clear all",
          empty: "Nothing selected",
        }}
        onClear={onClear}
      />,
    );

    expect(screen.getAllByText("2")).toHaveLength(2);
    fireEvent.click(screen.getByRole("button", { name: "Clear all" }));
    expect(onClear).toHaveBeenCalledOnce();
  });
});
