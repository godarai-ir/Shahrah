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

const users = [{
  id: "user-1",
  tenant_slug: "tenant-1",
  email: "member@example.com",
  role: "member",
  active: true,
  created_at: "2026-07-01T00:00:00Z",
  updated_at: "2026-07-01T00:00:00Z",
}];

function mockApi(userLimit: number, rejectInvite = false) {
  vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url === "/api/users/invite" && rejectInvite) {
      return { ok: false, status: 409, text: async () => JSON.stringify({ message: "Invite rejected" }) };
    }
    const payload = url === "/api/entitlements"
      ? { features: { "users.max": userLimit } }
      : url.startsWith("/api/roles")
        ? { data: [
            { id: "role-member", name: "member", display_name: "Member", is_system: true },
            { id: "role-support", name: "support_agent", display_name: "Support Agent", is_system: false },
            { id: "role-billing", name: "billing_manager", display_name: "Billing Manager", is_system: false },
          ] }
        : { data: users, total: users.length };
    return { ok: true, status: 200, text: async () => JSON.stringify(payload) };
  }));
}

describe("UsersAccessManagement", () => {
  beforeEach(() => {
    mockApi(2);
    Element.prototype.scrollIntoView = vi.fn();
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("opens the plan dialog when the users.max limit has been reached", async () => {
    mockApi(1);
    render(<UsersAccessManagement initialUsers={users} initialTotal={1} initialError={null} currentUserID="admin" canManageUsers />);

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.click(inviteButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("usersPage.upgradeDialog.description");
    expect(screen.queryByText("inviteTitle")).not.toBeInTheDocument();
  });

  it("opens the invite sheet from the left while capacity remains", async () => {
    render(<UsersAccessManagement initialUsers={users} initialTotal={1} initialError={null} currentUserID="admin" canManageUsers />);

    const inviteButton = screen.getByRole("button", { name: "inviteUser" });
    await waitFor(() => expect(inviteButton).toBeEnabled());
    fireEvent.focus(screen.getByText("usersPage.usage").closest("[tabindex]")!);
    expect((await screen.findAllByText("usersPage.limitTooltip")).length).toBeGreaterThan(0);
    fireEvent.click(inviteButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("inviteDescription");
    expect(screen.getByRole("dialog")).toHaveClass("left-0");

    fireEvent.click(screen.getByRole("combobox", { name: "form.role" }));
    expect(await screen.findByRole("option", { name: "Support Agent" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "roles.billingManager" })).toBeInTheDocument();
  });

  it("changes the role of selected users and reports the result", async () => {
    render(<UsersAccessManagement initialUsers={users} initialTotal={1} initialError={null} currentUserID="admin" canManageUsers />);

    await waitFor(() => expect(screen.getByRole("button", { name: "inviteUser" })).toBeEnabled());
    // Let the debounced first page load land; a new page clears the selection.
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

  it("validates invite email on the client before sending", async () => {
    render(<UsersAccessManagement initialUsers={users} initialTotal={1} initialError={null} currentUserID="admin" canManageUsers />);

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

  it("shows invite API errors inside the invite sheet", async () => {
    mockApi(2, true);
    render(<UsersAccessManagement initialUsers={users} initialTotal={1} initialError={null} currentUserID="admin" canManageUsers />);

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
