import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import VirtualUserTokenManagement from "./VirtualUserTokenManagement";

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
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

vi.mock("@/components/permissions-context", () => ({
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

    fireEvent.click(screen.getByRole("button", { name: "actions.deleteUser" }));

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
      if (url === "/api/entitlements") {
        return jsonResponse({ features: { "auth.vat.enabled": true } });
      }
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

    const createButtons = screen.getAllByText("actions.createUser");
    await waitFor(() => expect(createButtons[0].closest("button")).toBeEnabled());
    fireEvent.click(createButtons[0]);
    fireEvent.change(screen.getByLabelText("fields.alias"), { target: { value: "deploy-bot" } });
    fireEvent.change(screen.getByLabelText("fields.displayName"), { target: { value: "Deploy Bot" } });
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

  it("lets tenant admins enable on-behalf-of for a service token after confirmation", async () => {
    const serviceUser = {
      id: "vu-1",
      tenant_id: "tenant-1",
      alias: "librechat",
      team_id: "team-1",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/entitlements") {
        return jsonResponse({ features: { "auth.vat.enabled": true } });
      }
      if (url === "/api/virtual-users/vu-1/tokens/key-1/on-behalf-of" && init?.method === "PUT") {
        return jsonResponse({ enabled: true, product: "librechat" });
      }
      if (url.startsWith("/api/virtual-users/vu-1/tokens")) {
        return jsonResponse({ data: [{ id: "key-1", key_alias: "librechat-service", on_behalf_of: { enabled: false } }], total: 1 });
      }
      return jsonResponse({ data: [], total: 0 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[serviceUser]}
        initialTotal={1}
        initialError={null}
        teams={[{ id: "team-1", alias: "ops" }]}
        orgUnits={[]}
        canManageOnBehalfOf
      />,
    );

    fireEvent.click(await screen.findByRole("button", { name: "actions.enableOnBehalfOf" }));
    expect(screen.getByText("confirm.enableOnBehalfOfTitle")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "confirm.confirm" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/virtual-users/vu-1/tokens/key-1/on-behalf-of",
        expect.objectContaining({ method: "PUT", body: JSON.stringify({ enabled: true }) }),
      );
    });
  });

  it("hides the on-behalf-of action from non-admins and shows the enabled badge", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/entitlements") {
        return jsonResponse({ features: { "auth.vat.enabled": true } });
      }
      if (url.startsWith("/api/virtual-users/vu-1/tokens")) {
        return jsonResponse({ data: [{ id: "key-1", key_alias: "librechat-service", on_behalf_of: { enabled: true, product: "librechat" } }], total: 1 });
      }
      return jsonResponse({ data: [], total: 0 });
    }));

    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[{
          id: "vu-1",
          tenant_id: "tenant-1",
          alias: "librechat",
          team_id: "team-1",
          active: true,
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        }]}
        initialTotal={1}
        initialError={null}
        teams={[{ id: "team-1", alias: "ops" }]}
        orgUnits={[]}
      />,
    );

    expect(await screen.findByText("status.onBehalfOf")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "actions.disableOnBehalfOf" })).not.toBeInTheDocument();
  });

  it("shows the upgrade dialog when VAT is disabled by the commercial plan", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      if (String(input) === "/api/entitlements") {
        return jsonResponse({ features: { "auth.vat.enabled": false } });
      }
      return jsonResponse({ data: [], total: 0 });
    }));

    render(
      <VirtualUserTokenManagement
        initialVirtualUsers={[]}
        initialTotal={0}
        initialError={null}
        teams={[]}
        orgUnits={[]}
      />,
    );

    const createButtons = screen.getAllByText("actions.createUser");
    await waitFor(() => expect(createButtons[0].closest("button")).toBeEnabled());
    fireEvent.click(createButtons[0]);

    expect(await screen.findByText("upgradeDialog.title")).toBeInTheDocument();
    expect(screen.getByText("upgradeDialog.action")).toBeInTheDocument();
  });
});

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
