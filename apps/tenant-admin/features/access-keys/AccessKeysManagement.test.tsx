import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import AccessKeysManagement from "./AccessKeysManagement";
import type { TenantKey } from "./api";

const tMock = (key: string, values?: Record<string, unknown>) => {
  if (!values) return key;
  return Object.entries(values).reduce(
    (text, [name, value]) => text.replace(`{${name}}`, String(value)),
    key,
  );
};

vi.mock("next-intl", () => ({ useTranslations: () => tMock, useLocale: () => "en" }));

const members = [
  {
    id: "user-1",
    tenant_slug: "acme",
    email: "ada@example.com",
    role: "member",
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const keys: TenantKey[] = [
  { id: "key-1", key_alias: "ada-cli", key_type: "pat", created_by_user_id: "user-1" },
  { id: "key-2", key_alias: "billing-bot", key_type: "vat", virtual_user_id: "vu-1" },
  { id: "key-3", key_alias: "other-cli", key_type: "pat", created_by_user_id: "user-2" },
];

function renderKeys(overrides: Partial<React.ComponentProps<typeof AccessKeysManagement>> = {}) {
  return render(
    <AccessKeysManagement
      initialKeys={keys}
      initialError={null}
      members={members}
      virtualUsers={[{
        id: "vu-1",
        tenant_id: "t1",
        alias: "billing-bot",
        display_name: "Billing service",
        active: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      }]}
      currentUserID="user-1"
      isTenantWide
      {...overrides}
    />,
  );
}

describe("AccessKeysManagement", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, text: async () => JSON.stringify({ data: keys }) })));
    Element.prototype.scrollIntoView = vi.fn();
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("lists personal and service keys together with a type column", () => {
    renderKeys();

    expect(screen.getByText("ada-cli")).toBeInTheDocument();
    expect(screen.getByText("billing-bot")).toBeInTheDocument();
    // The owner column names the service, rather than repeating the key alias.
    expect(screen.getByText("Billing service")).toBeInTheDocument();
    expect(screen.getAllByText("keys.typePat")).toHaveLength(2);
    expect(screen.getByText("keys.typeVat")).toBeInTheDocument();
  });

  it("names the member who owns a personal key", () => {
    renderKeys();

    expect(screen.getByText("ada")).toBeInTheDocument();
  });

  it("filters by key type", async () => {
    renderKeys();

    fireEvent.click(screen.getByRole("combobox", { name: "keys.filterType" }));
    fireEvent.click(await screen.findByRole("option", { name: "keys.typeVat" }));

    await waitFor(() => expect(screen.queryByText("ada-cli")).not.toBeInTheDocument());
    expect(screen.getByText("billing-bot")).toBeInTheDocument();
  });

  it("lets a member revoke their own key but not someone else's", () => {
    renderKeys();

    const ownRow = screen.getByText("ada-cli").closest("tr")!;
    expect(within(ownRow).getByRole("button", { name: "actions.revoke" })).toBeEnabled();

    const otherRow = screen.getByText("other-cli").closest("tr")!;
    expect(within(otherRow).getByRole("button", { name: "actions.revoke" })).toBeDisabled();
  });

  it("hides the owner column when the caller only sees their own keys", () => {
    renderKeys({ isTenantWide: false, initialKeys: [keys[0]] });

    expect(screen.queryByRole("columnheader", { name: "keys.owner" })).not.toBeInTheDocument();
    expect(screen.getByText("ada-cli")).toBeInTheDocument();
  });

  it("marks an expired key in the expiry column", () => {
    renderKeys({ initialKeys: [{ id: "key-x", key_alias: "old", key_type: "pat", expires_at: "2020-01-01T00:00:00Z" }] });

    expect(screen.getByText("keys.state.expired")).toBeInTheDocument();
  });

  it("shows no-expiry keys as such", () => {
    renderKeys({ initialKeys: [keys[0]] });

    expect(screen.getByText("keys.noExpiry")).toBeInTheDocument();
  });
});
