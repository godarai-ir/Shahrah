import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import VirtualUserTokenManagement from "./virtual-user-token-management";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string, values?: Record<string, unknown>) => {
    if (!values) return key;
    return Object.entries(values).reduce(
      (text, [name, value]) => text.replace(`{${name}}`, String(value)),
      key,
    );
  },
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/virtual-users",
}));

vi.mock("@/app/components/permissions-context", () => ({
  usePermissions: () => ["virtual_account:read", "virtual_account:manage", "virtual_account:create"],
}));

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("VirtualUserTokenManagement", () => {
  it("renders the virtual user empty state", () => {
    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[]}
        initialTotal={0}
        initialError={null}
        teams={[]}
        orgUnits={[]}
      />,
    );

    expect(screen.getByText("usersTable.empty")).toBeInTheDocument();
    expect(screen.getByText("detail.emptyTitle")).toBeInTheDocument();
  });

  it("opens delete confirmation for a selected virtual user", () => {
    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[
          {
            id: "vu-1",
            tenant_id: "tenant-1",
            alias: "deploy-bot",
            display_name: "Deploy Bot",
            team_id: "team-1",
            active: true,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        ]}
        initialTotal={1}
        initialError={null}
        teams={[{ id: "team-1", alias: "ops" }]}
        orgUnits={[]}
      />,
    );

    fireEvent.click(screen.getByText("actions.deleteUser"));

    expect(screen.getByText("deleteUser.title")).toBeInTheDocument();
    expect(screen.getByText("deleteUser.description")).toBeInTheDocument();
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("loads allowed models for a newly created virtual user without a page refresh", async () => {
    const createdUser = {
      id: "vu-created",
      tenant_id: "tenant-1",
      alias: "deploy-bot",
      display_name: "Deploy Bot",
      team_id: "team-1",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/virtual-users" && init?.method === "POST") {
        return jsonResponse(createdUser);
      }
      if (url.startsWith("/api/virtual-users/vu-created/tokens")) {
        return jsonResponse({ data: [], total: 0 });
      }
      if (url.startsWith("/api/allowed-models?key_type=vat&team_id=team-1")) {
        return jsonResponse({ data: [{ id: "gpt-4.1-mini" }] });
      }
      if (url.startsWith("/api/virtual-users?")) {
        return jsonResponse({ data: [createdUser], total: 1 });
      }
      return jsonResponse({ data: [], total: 0 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[]}
        initialTotal={0}
        initialError={null}
        teams={[{ id: "team-1", alias: "ops" }]}
        orgUnits={[]}
      />,
    );

    fireEvent.click(screen.getByText("actions.createUser"));
    fireEvent.change(screen.getByLabelText("fields.alias"), { target: { value: "deploy-bot" } });
    fireEvent.change(screen.getByLabelText("fields.displayName"), { target: { value: "Deploy Bot" } });
    fireEvent.change(screen.getByLabelText("fields.team"), { target: { value: "team-1" } });
    fireEvent.click(screen.getAllByText("actions.createUser").at(-1)!);

    expect(await screen.findAllByText("Deploy Bot")).toHaveLength(2);

    fireEvent.click(screen.getByText("actions.createToken"));

    expect(await screen.findByText("gpt-4.1-mini")).toBeInTheDocument();
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/allowed-models?key_type=vat&team_id=team-1",
        expect.objectContaining({ cache: "no-store" }),
      );
    });
  });
});

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
