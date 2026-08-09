import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import RolesPermissionsManagement from "./RolesPermissionsManagement";

const pushMock = vi.fn();
const refreshMock = vi.fn();
let permissionsMock = ["role:list", "role:manage"];

const tMock = (key: string, values?: Record<string, unknown>) => {
  if (!values) return key;
  return Object.entries(values).reduce(
    (text, [name, value]) => text.replace(`{${name}}`, String(value)),
    key,
  );
};

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
  useTranslations: () => tMock,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, refresh: refreshMock }),
}));

vi.mock("@/components/permissions-context", () => ({
  usePermissions: () => permissionsMock,
}));

const roles = [
  {
    id: "role-1",
    tenant_id: "tenant-1",
    name: "billing_manager",
    display_name: "Billing Manager",
    description: "Manages billing",
    is_system: true,
    permissions: ["billing:read", "billing:manage"],
    created_at: "2026-07-01T00:00:00Z",
    updated_at: "2026-07-01T00:00:00Z",
  },
  {
    id: "role-2",
    tenant_id: "tenant-1",
    name: "support_agent",
    display_name: "Support Agent",
    description: "Answers support requests",
    is_system: false,
    permissions: ["support:read"],
    created_at: "2026-07-01T00:00:00Z",
    updated_at: "2026-07-01T00:00:00Z",
  },
];

const permissionCatalog = [
  { key: "role:list", group: "roles", description: "List roles" },
];

function renderManagement() {
  return render(
    <RolesPermissionsManagement
      initialRoles={roles}
      initialError={null}
      initialPermissionCatalog={permissionCatalog}
    />,
  );
}

describe("RolesPermissionsManagement", () => {
  beforeEach(() => {
    permissionsMock = ["role:list", "role:manage"];
    pushMock.mockReset();
    refreshMock.mockReset();
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ features: { "rbac.enabled": true } }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    )));
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the role list without duplicating the route heading", () => {
    renderManagement();

    expect(screen.queryByRole("heading", { level: 1 })).not.toBeInTheDocument();
    expect(screen.getByText("billing_manager")).toBeInTheDocument();
    expect(screen.getByText("support_agent")).toBeInTheDocument();
  });

  it("places the create action beside the role type filter", () => {
    renderManagement();

    const roleTypeFilter = screen.getByRole("combobox", { name: "rolesPage.filters.typeLabel" });
    const createButton = screen.getByRole("button", { name: "rolesPage.form.create" });

    expect(roleTypeFilter.parentElement?.parentElement).toContainElement(createButton);
  });

  it("filters roles by their searchable fields", () => {
    renderManagement();

    fireEvent.change(screen.getByRole("textbox", { name: "rolesPage.filters.searchLabel" }), {
      target: { value: "support" },
    });

    expect(screen.queryByText("billing_manager")).not.toBeInTheDocument();
    expect(screen.getByText("support_agent")).toBeInTheDocument();
  });

  it("omits the actions column for read-only users", () => {
    permissionsMock = ["role:list"];

    renderManagement();

    expect(screen.queryByText("rolesPage.table.actions")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /rolesPage\.actions\.openMenu/ })).not.toBeInTheDocument();
  });

  it("offers retry without showing stale table content after a load error", () => {
    render(<RolesPermissionsManagement initialRoles={roles} initialError="Request failed" initialPermissionCatalog={permissionCatalog} />);

    expect(screen.getByRole("alert")).toHaveTextContent("Request failed");
    expect(screen.queryByRole("table")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "rolesPage.actions.retry" }));
    expect(refreshMock).toHaveBeenCalledOnce();
  });

  it("opens the upgrade dialog instead of navigating when RBAC is unavailable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ features: { "rbac.enabled": false } }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    )));

    renderManagement();

    const createButton = screen.getByRole("button", { name: "rolesPage.form.create" });
    await waitFor(() => expect(createButton).toBeEnabled());
    fireEvent.click(createButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("rolesPage.upgradeDialog.description");
    expect(pushMock).not.toHaveBeenCalled();
  });

  it("opens role creation in a left-side drawer when RBAC is enabled", async () => {
    renderManagement();

    const createButton = screen.getByRole("button", { name: "rolesPage.form.create" });
    await waitFor(() => expect(createButton).toBeEnabled());
    fireEvent.click(createButton);

    expect(screen.getByRole("dialog", { name: "rolesPage.createTitle" })).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });
});
