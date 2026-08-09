import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import RoleEditManagement from "./RolesEditManagement";

const pushMock = vi.fn();
const refreshMock = vi.fn();

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
  useTranslations: () => (key: string) => key,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, refresh: refreshMock }),
}));

const role = {
  id: "role-1",
  tenant_id: "tenant-1",
  name: "org_admin",
  display_name: "Org Admin",
  description: "System role",
  is_system: true,
  permissions: ["role:list"],
  created_at: "2026-07-01T00:00:00Z",
  updated_at: "2026-07-01T00:00:00Z",
};

const permissionCatalog = [
  { key: "role:list", group: "roles", description: "List roles" },
  { key: "team:read", group: "teams", description: "Read teams" },
];

describe("RoleEditManagement", () => {
  beforeEach(() => {
    pushMock.mockReset();
    refreshMock.mockReset();
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, text: async () => "{}" }));
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders an accessible role form with a single page heading", () => {
    render(<RoleEditManagement initialRole={role} initialPermissionCatalog={permissionCatalog} />);

    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    const roleSlug = screen.getByRole("textbox", { name: "شناسه فنی" });
    const displayName = screen.getByRole("textbox", { name: "rolesPage.form.displayName" });
    expect(roleSlug).toHaveValue("org_admin");
    expect(roleSlug).toHaveAttribute("dir", "ltr");
    expect(roleSlug).toHaveAttribute("pattern", "[A-Za-z0-9_-]+");
    expect(displayName).toHaveValue("Org Admin");
    expect(displayName).toHaveAttribute("dir", "ltr");
    expect(screen.getByRole("textbox", { name: "rolesPage.form.description" })).toHaveValue("System role");
  });

  it("shows the selected permission and exposes both permission views", () => {
    render(<RoleEditManagement initialRole={role} initialPermissionCatalog={permissionCatalog} />);

    expect(screen.getByRole("tab", { name: "انتخاب‌شده‌ها" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: "همه مجوزها" })).toBeInTheDocument();

    const rolePermission = screen.getByRole("checkbox", { name: /اجازه مشاهده فهرست نقش/ });
    expect(rolePermission).toBeChecked();
  });

  it("submits the edited role and returns to the role list", async () => {
    render(<RoleEditManagement initialRole={role} initialPermissionCatalog={permissionCatalog} />);

    fireEvent.change(screen.getByRole("textbox", { name: "rolesPage.form.displayName" }), {
      target: { value: "مدیر سازمان" },
    });
    expect(screen.getByRole("textbox", { name: "rolesPage.form.displayName" })).toHaveAttribute("dir", "rtl");
    fireEvent.click(screen.getByRole("button", { name: "rolesPage.actions.save" }));

    await waitFor(() => expect(fetch).toHaveBeenCalledOnce());
    expect(fetch).toHaveBeenCalledWith(
      "/api/roles/role-1",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({
          display_name: "مدیر سازمان",
          description: "System role",
          permissions: ["role:list"],
        }),
      }),
    );
    expect(pushMock).toHaveBeenCalledWith("/roles");
    expect(refreshMock).toHaveBeenCalledOnce();
  });
});
