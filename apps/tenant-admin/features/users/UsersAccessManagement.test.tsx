import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import UsersAccessManagement from "./UsersAccessManagement";

const tMock = (key: string, values?: Record<string, unknown>) => {
  if (!values) return key;
  return Object.entries(values).reduce(
    (text, [name, value]) => text.replace(`{${name}}`, String(value)),
    key,
  );
};

vi.mock("next-intl", () => ({ useTranslations: () => tMock, useLocale: () => "en" }));

const replaceMock = vi.fn();
let currentSearch = "";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock, push: vi.fn(), refresh: vi.fn() }),
  usePathname: () => "/users",
  useSearchParams: () => new URLSearchParams(currentSearch),
}));

const users = [{
  id: "user-1",
  tenant_slug: "tenant-1",
  email: "member@example.com",
  role: "member",
  roles: ["member", "billing_manager"],
  active: true,
  created_at: "2026-07-01T00:00:00Z",
  updated_at: "2026-07-01T00:00:00Z",
}];

const roles = [
  { id: "role-member", name: "member", display_name: "Member", is_system: true, permissions: ["user:list"] },
  { id: "role-support", name: "support_agent", display_name: "Support Agent", is_system: false, permissions: [] },
  { id: "role-billing", name: "billing_manager", display_name: "Billing Manager", is_system: false, permissions: ["billing:read"] },
];

function mockApi(userLimit: number, rejectInvite: false | "conflict" | "plan_limit" = false, usedSeats = users.length) {
  vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url === "/api/users/invite" && rejectInvite === "conflict") {
      return { ok: false, status: 409, text: async () => JSON.stringify({ message: "Invite rejected" }) };
    }
    if (url === "/api/users/invite" && rejectInvite === "plan_limit") {
      return {
        ok: false,
        status: 403,
        text: async () => JSON.stringify({ error: { code: "plan_user_limit_reached", message: "The plan's user limit is reached." } }),
      };
    }
    const payload = url === "/api/users/seats"
      ? { used: usedSeats, active_members: users.length, pending_invitations: usedSeats - users.length, limit: userLimit }
      : url === "/api/access-keys"
        ? { data: [{ id: "key-1", key_alias: "cli", key_type: "pat", created_by_user_id: "user-1" }] }
        : url.startsWith("/api/roles")
          ? { data: roles }
          : { data: users, total: users.length };
    return { ok: true, status: 200, text: async () => JSON.stringify(payload) };
  }));
}

function renderPage() {
  return render(
    <UsersAccessManagement
      initialUsers={users}
      initialError={null}
      initialPermissionCatalog={[{ key: "billing:read", group: "billing", description: "" }]}
      currentUserID="admin"
      canManageUsers
    />,
  );
}

describe("UsersAccessManagement", () => {
  beforeEach(() => {
    currentSearch = "";
    replaceMock.mockClear();
    mockApi(2);
    Element.prototype.scrollIntoView = vi.fn();
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("opens the plan dialog when the users.max limit has been reached", async () => {
    mockApi(1);
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("usersPage.upgradeDialog.description");
    expect(screen.queryByText("inviteTitle")).not.toBeInTheDocument();
  });

  it("opens the invite sheet from the inline-end edge while capacity remains", async () => {
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("inviteDescription");
    expect(screen.getByRole("dialog")).toHaveClass("end-0");

    fireEvent.click(screen.getByRole("combobox", { name: "form.role" }));
    expect(await screen.findByRole("option", { name: "Support Agent" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "roles.billingManager" })).toBeInTheDocument();
  });

  it("shows every role on a member, not just the first", async () => {
    renderPage();

    await waitFor(() => expect(screen.getByText("roles.member")).toBeInTheDocument());
    // First role as a badge, the rest folded into a "+N" chip.
    expect(screen.getByText("table.moreRoles")).toBeInTheDocument();
  });

  it("writes the role filter to the URL instead of local state", async () => {
    renderPage();

    await waitFor(() => expect(screen.getByRole("combobox", { name: "filters.role" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("combobox", { name: "filters.role" }));
    fireEvent.click(await screen.findByRole("option", { name: "roles.billingManager" }));

    expect(replaceMock).toHaveBeenCalledWith("/users?role=billing_manager", { scroll: false });
  });

  it("filters the roster by status without refetching", async () => {
    currentSearch = "status=deactivated";
    renderPage();

    await waitFor(() => expect(vi.mocked(fetch).mock.calls.some(([url]) => String(url).startsWith("/api/users?"))).toBe(true));
    // The only member is active, so a deactivated filter empties the table.
    await waitFor(() => expect(screen.getByText("table.noUsersFound")).toBeInTheDocument());
  });

  it("opens the member drawer from the URL and lists effective permissions", async () => {
    currentSearch = "member=user-1";
    renderPage();

    const drawer = await screen.findByRole("dialog");
    expect(drawer).toHaveTextContent("member.effectivePerms");
    expect(drawer).toHaveTextContent("billing:read");
    expect(drawer).toHaveTextContent("perms.grantCount");
  });

  it("changes the role of selected users and reports the result", async () => {
    renderPage();

    await waitFor(() => expect(screen.getByRole("button", { name: "inviteUser" })).toBeEnabled());
    await waitFor(() => expect(vi.mocked(fetch).mock.calls.some(([url]) => String(url).startsWith("/api/users?"))).toBe(true));
    await new Promise((resolve) => setTimeout(resolve, 0));

    const row = screen.getByRole("cell", { name: "member@example.com" }).closest("tr")!;
    fireEvent.click(within(row).getByRole("checkbox", { name: "table.selectRow" }));
    const bulkBar = screen.getByRole("toolbar", { name: "table.bulkActions" });
    fireEvent.click(within(bulkBar).getByRole("button", { name: "actions.changeRole" }));

    const dialog = screen.getByRole("dialog");
    fireEvent.click(within(dialog).getByRole("radio", { name: /Support Agent/ }));
    fireEvent.click(within(dialog).getByRole("button", { name: "roleForm.submit" }));

    expect(await screen.findByRole("status")).toHaveTextContent("bulk.result");
    expect(fetch).toHaveBeenCalledWith("/api/users/roles", expect.objectContaining({ method: "PATCH" }));
    expect(screen.queryByRole("toolbar", { name: "table.bulkActions" })).not.toBeInTheDocument();
  });

  it("deactivates selected users in bulk", async () => {
    renderPage();

    await waitFor(() => expect(vi.mocked(fetch).mock.calls.some(([url]) => String(url).startsWith("/api/users?"))).toBe(true));
    await new Promise((resolve) => setTimeout(resolve, 0));

    const row = screen.getByRole("cell", { name: "member@example.com" }).closest("tr")!;
    fireEvent.click(within(row).getByRole("checkbox", { name: "table.selectRow" }));
    const bulkBar = screen.getByRole("toolbar", { name: "table.bulkActions" });
    fireEvent.click(within(bulkBar).getByRole("button", { name: "actions.deactivate" }));

    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith("/api/users/user-1/deactivate", expect.objectContaining({ method: "POST" })),
    );
  });

  it("validates invite email on the client before sending", async () => {
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);
    const dialog = screen.getByRole("dialog");
    const emailInput = within(dialog).getByRole("textbox", { name: "form.email" });
    fireEvent.change(emailInput, { target: { value: "invalid-email" } });
    fireEvent.blur(emailInput);

    expect(within(dialog).getByRole("alert")).toHaveTextContent("messages.invalidEmail");
    expect(emailInput).toHaveAttribute("dir", "ltr");
  });

  it("counts pending invitations as seats before offering the invite sheet", async () => {
    mockApi(2, false, 2);
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("usersPage.upgradeDialog.description");
  });

  it("explains a full plan when the server refuses an invite", async () => {
    mockApi(5, "plan_limit");
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);
    const dialog = screen.getByRole("dialog");
    fireEvent.change(within(dialog).getByRole("textbox", { name: "form.email" }), { target: { value: "new@example.com" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "button.createInvite" }));

    expect(await within(dialog).findByRole("alert")).toHaveTextContent("messages.planUserLimitReached");
  });

  it("shows invite API errors inside the invite sheet", async () => {
    mockApi(2, "conflict");
    renderPage();

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);
    const dialog = screen.getByRole("dialog");
    fireEvent.change(within(dialog).getByRole("textbox", { name: "form.email" }), { target: { value: "new@example.com" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "button.createInvite" }));

    expect(await within(dialog).findByRole("alert")).toHaveTextContent("Invite rejected");
    expect(screen.getAllByRole("alert")).toHaveLength(1);
  });
});
