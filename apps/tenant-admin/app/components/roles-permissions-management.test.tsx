import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import RolesPermissionsManagement from "./roles-permissions-management";

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

vi.mock("@/app/components/permissions-context", () => ({
  usePermissions: () => permissionsMock,
}));

vi.mock("@/app/components/access-management-header", () => ({
  default: () => <nav aria-label="access-navigation" />,
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

describe("RolesPermissionsManagement", () => {
  beforeEach(() => {
    permissionsMock = ["role:list", "role:manage"];
    pushMock.mockReset();
    refreshMock.mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders one page heading and the role list", () => {
    render(<RolesPermissionsManagement initialRoles={roles} initialError={null} />);

    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByText("billing_manager")).toBeInTheDocument();
    expect(screen.getByText("support_agent")).toBeInTheDocument();
  });

  it("filters roles by their searchable fields", () => {
    render(<RolesPermissionsManagement initialRoles={roles} initialError={null} />);

    fireEvent.change(screen.getByRole("textbox", { name: "rolesPage.filters.searchLabel" }), {
      target: { value: "support" },
    });

    expect(screen.queryByText("billing_manager")).not.toBeInTheDocument();
    expect(screen.getByText("support_agent")).toBeInTheDocument();
  });

  it("omits the actions column for read-only users", () => {
    permissionsMock = ["role:list"];

    render(<RolesPermissionsManagement initialRoles={roles} initialError={null} />);

    expect(screen.queryByText("rolesPage.table.actions")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /rolesPage\.actions\.openMenu/ })).not.toBeInTheDocument();
  });

  it("offers retry without showing stale table content after a load error", () => {
    render(<RolesPermissionsManagement initialRoles={roles} initialError="Request failed" />);

    expect(screen.getByRole("alert")).toHaveTextContent("Request failed");
    expect(screen.queryByRole("table")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "rolesPage.actions.retry" }));
    expect(refreshMock).toHaveBeenCalledOnce();
  });
});
