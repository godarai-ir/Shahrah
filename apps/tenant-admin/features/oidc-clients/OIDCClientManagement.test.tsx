import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import OIDCClientManagement from "./OIDCClientManagement";
import type { OIDCClient } from "./api";

const tMock = (key: string, values?: Record<string, unknown>) => {
  if (!values) return key;
  return Object.entries(values).reduce((text, [name, value]) => text.replace(`{${name}}`, String(value)), key);
};

vi.mock("next-intl", () => ({ useTranslations: () => tMock, useLocale: () => "en" }));

const toastMock = vi.hoisted(() => vi.fn());
vi.mock("@/components/app-ui/feedback-toast", () => ({ showFeedbackToast: toastMock }));

const libreChat: OIDCClient = {
  id: "c-1",
  client_id: "gdc-0123456789abcdef0123456789abcdef",
  name: "LibreChat",
  redirect_uris: ["https://chat.acme.example/oauth/openid/callback", "https://chat2.acme.example/oauth/openid/callback"],
  post_logout_redirect_uris: [],
  active: true,
  created_at: "2026-10-07T09:00:00Z",
  secret_rotated_at: "2026-10-07T09:00:00Z",
};

type FetchCall = { url: string; method: string; body: unknown };
let calls: FetchCall[] = [];
let responses: Array<{ status: number; body: unknown }> = [];

function respondWith(...next: Array<{ status: number; body: unknown }>) {
  responses.push(...next);
}

function renderPage(overrides: Partial<React.ComponentProps<typeof OIDCClientManagement>> = {}) {
  return render(
    <OIDCClientManagement
      initialClients={[libreChat]}
      initialError={null}
      issuerURL="https://acme.godar.example"
      {...overrides}
    />,
  );
}

function openRowMenu(name: string) {
  fireEvent.pointerDown(screen.getByRole("button", { name: `actions.menu`.replace("{name}", name) }), { button: 0, ctrlKey: false });
}

describe("OIDCClientManagement", () => {
  beforeEach(() => {
    calls = [];
    responses = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) => {
        calls.push({ url, method: init?.method ?? "GET", body: init?.body ? JSON.parse(String(init.body)) : undefined });
        const next = responses.shift() ?? { status: 200, body: { data: [libreChat] } };
        return { ok: next.status < 400, status: next.status, text: async () => JSON.stringify(next.body) };
      }),
    );
    Element.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    toastMock.mockReset();
  });

  it("lists clients next to the integration values a relying party needs", () => {
    renderPage();

    expect(screen.getByText("LibreChat")).toBeInTheDocument();
    expect(screen.getByText(libreChat.client_id)).toBeInTheDocument();
    expect(screen.getByText("table.moreURIs")).toBeInTheDocument();
    expect(screen.getByDisplayValue("https://acme.godar.example")).toBeInTheDocument();
    expect(screen.getByDisplayValue("https://acme.godar.example/.well-known/openid-configuration")).toBeInTheDocument();
    expect(screen.getByDisplayValue("https://acme.godar.example/v1")).toBeInTheDocument();
  });

  it("registers a client and reveals its secret exactly once", async () => {
    respondWith(
      { status: 201, body: { ...libreChat, id: "c-2", name: "Docs bot", client_id: "gdc-new", client_secret: "gds_secret_value" } },
      { status: 200, body: { data: [libreChat] } },
    );
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "actions.create" }));
    fireEvent.change(await screen.findByLabelText("fields.name"), { target: { value: "Docs bot" } });
    fireEvent.change(screen.getByLabelText("fields.redirectURIs"), {
      target: { value: "https://docs.acme.example/cb\n\nhttps://docs.acme.example/cb" },
    });
    fireEvent.submit(document.getElementById("oidc-client-form")!);

    expect(await screen.findByDisplayValue("gds_secret_value")).toBeInTheDocument();
    expect(calls[0]).toEqual({
      url: "/api/oidc-clients",
      method: "POST",
      body: {
        name: "Docs bot",
        redirect_uris: ["https://docs.acme.example/cb"],
        post_logout_redirect_uris: [],
        credit_product: "oidc_app",
      },
    });

    fireEvent.click(screen.getAllByRole("button", { name: "actions.close" }).at(-1)!);
    await waitFor(() => expect(screen.queryByDisplayValue("gds_secret_value")).not.toBeInTheDocument());
  });

  it("rejects an insecure redirect URI before calling the API", async () => {
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "actions.create" }));
    fireEvent.change(await screen.findByLabelText("fields.name"), { target: { value: "Bad" } });
    fireEvent.change(screen.getByLabelText("fields.redirectURIs"), { target: { value: "http://chat.example.com/cb" } });
    fireEvent.submit(document.getElementById("oidc-client-form")!);

    expect(await screen.findByText("errors.redirect_invalid")).toBeInTheDocument();
    expect(calls).toHaveLength(0);
  });

  it("maps a control-plane error code to localized copy and keeps the sheet open", async () => {
    respondWith({ status: 422, body: { error: { code: "oidc_client_invalid", message: "raw upstream text" } } });
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "actions.create" }));
    fireEvent.change(await screen.findByLabelText("fields.name"), { target: { value: "App" } });
    fireEvent.change(screen.getByLabelText("fields.redirectURIs"), { target: { value: "https://app.example/cb" } });
    fireEvent.submit(document.getElementById("oidc-client-form")!);

    expect(await screen.findByText("apiErrors.oidc_client_invalid")).toBeInTheDocument();
    expect(screen.queryByText("raw upstream text")).not.toBeInTheDocument();
    expect(screen.getByLabelText("fields.name")).toBeInTheDocument();
  });

  it("deactivates a client from the row menu", async () => {
    respondWith({ status: 200, body: { ...libreChat, active: false } }, { status: 200, body: { data: [{ ...libreChat, active: false }] } });
    renderPage();

    openRowMenu("LibreChat");
    fireEvent.click(screen.getByRole("menuitem", { name: "actions.deactivate" }));

    await waitFor(() => expect(screen.getByText("status.inactive")).toBeInTheDocument());
    expect(calls[0]).toEqual({ url: "/api/oidc-clients/c-1", method: "PATCH", body: { active: false } });
  });

  it("rotates the secret after confirmation and shows the new one", async () => {
    respondWith({ status: 200, body: { ...libreChat, client_secret: "gds_rotated" } });
    renderPage();

    openRowMenu("LibreChat");
    fireEvent.click(screen.getByRole("menuitem", { name: "actions.rotateSecret" }));
    fireEvent.click(await screen.findByRole("button", { name: "rotateDialog.confirm" }));

    expect(await screen.findByDisplayValue("gds_rotated")).toBeInTheDocument();
    expect(screen.getByText("secret.rotatedTitle")).toBeInTheDocument();
    expect(calls[0]).toMatchObject({ url: "/api/oidc-clients/c-1/rotate-secret", method: "POST" });
  });

  it("only deletes after the client name is typed", async () => {
    respondWith({ status: 204, body: {} }, { status: 200, body: { data: [] } });
    renderPage();

    openRowMenu("LibreChat");
    fireEvent.click(screen.getByRole("menuitem", { name: "actions.delete" }));
    const dialog = await screen.findByRole("dialog");
    const confirm = within(dialog).getByRole("button", { name: "deleteDialog.confirm" });
    expect(confirm).toBeDisabled();

    fireEvent.change(within(dialog).getByRole("textbox"), { target: { value: "LibreChat" } });
    fireEvent.click(confirm);

    await waitFor(() => expect(calls[0]).toMatchObject({ url: "/api/oidc-clients/c-1", method: "DELETE" }));
    expect(await screen.findByText("table.empty")).toBeInTheDocument();
  });

  it("explains the access requirement instead of an empty list for non-admins", () => {
    renderPage({ initialClients: [], initialError: "forbidden" });

    expect(screen.getByText("forbidden.title")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "actions.create" })).not.toBeInTheDocument();
  });

  it("surfaces a load failure with a retryable notice", () => {
    renderPage({ initialClients: [], initialError: "unavailable" });

    expect(screen.getByText("messages.loadError")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "actions.create" })).toBeInTheDocument();
  });
});
