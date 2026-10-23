import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { SeatRenewalPreview, SeatSummary } from "./api";
import { RenewSeatsForm } from "./RenewSeatsForm";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    billing: (await import("@/messages/modules/billing/en.json")).default,
    billingSeats: (await import("@/messages/modules/billingSeats/en.json")).default,
  };
  const translators = new Map<string, unknown>();
  return {
    ...actual,
    useLocale: () => "en",
    useTranslations: (namespace: keyof typeof messages) => {
      if (!translators.has(namespace)) translators.set(namespace, actual.createTranslator({ locale: "en", messages, namespace }));
      return translators.get(namespace);
    },
  };
});

vi.mock("@/hooks/use-locale-format", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/hooks/use-locale-format")>()),
  useAppLocale: () => "en",
}));

const baseSummary: SeatSummary = {
  enforced: true,
  subscription_status: "active",
  subscription_kind: "seats",
  billing_interval: "monthly",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  renewal_opens_at: "2026-10-25T00:00:00Z",
  free_limit: 5,
  free_used: 1,
  unassigned: 0,
  pending_invitations: 0,
  paid: [
    { type: "pro", purchased: 3, assigned: 3, available: 0, renewing: 3 },
    { type: "max", purchased: 0, assigned: 0, available: 0, renewing: 0 },
  ],
  offers: [],
  pending_requests: 0,
  renewal_due: {
    lines: [{ block_id: "b1", seat: "pro", quantity: 3, paid_through: "2026-11-01T00:00:00Z" }],
    period_end: "2026-12-01T00:00:00Z",
    price_irr: 54_000_000,
    tax_percent: "10.00",
    tax_irr: 5_400_000,
    payable_irr: 59_400_000,
  },
  tax_percent: "10.00",
};

function previewFor(keep: number): SeatRenewalPreview {
  const losing = ["Ali", "Sara", "Reza"].slice(keep).map((name) => ({
    user_id: `u-${name}`,
    email: `${name.toLowerCase()}@example.com`,
    name,
    assigned_at: "2026-09-10T00:00:00Z",
    fallback: name === "Reza" ? ("free" as const) : ("unassigned" as const),
  }));

  return {
    period_end: "2026-11-01T00:00:00Z",
    next_period_end: "2026-12-01T00:00:00Z",
    lines: [{ seat: "pro", held: 3, renewed: 0, keep, assigned: 3, unit_price_irr: 18_000_000, losing }],
    price_irr: 18_000_000 * keep,
    tax_percent: "10.00",
    tax_irr: 1_800_000 * keep,
    payable_irr: 19_800_000 * keep,
  };
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

const insideWindow = Date.parse("2026-10-28T00:00:00Z");
const beforeWindow = Date.parse("2026-10-10T00:00:00Z");

function renderForm(props: Partial<React.ComponentProps<typeof RenewSeatsForm>> = {}) {
  return render(<RenewSeatsForm summary={baseSummary} preview={previewFor(3)} windowOpen canManage paymentResult={null} nowMs={insideWindow} {...props} />);
}

describe("RenewSeatsForm", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("is an empty state with a way to buy when there is nothing to renew", () => {
    renderForm({ summary: { ...baseSummary, subscription_kind: "free", period_end: null, paid: baseSummary.paid.map((item) => ({ ...item, purchased: 0 })) }, preview: null });

    expect(screen.getByText("No paid seats to renew")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Buy seats" })).toHaveAttribute("href", "/billing/seats/buy");
  });

  it("says renewal is done and asks for nothing when every seat is paid", () => {
    renderForm({ summary: { ...baseSummary, renewal_due: null } });

    expect(screen.getByText("Renewal is done")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Pay/ })).not.toBeInTheDocument();
  });

  it("shows how many seats stay and that nobody loses a seat when all are kept", () => {
    renderForm();

    expect(screen.getByLabelText("Pro seats to renew")).toHaveValue(3);
    expect(screen.getAllByText("No member loses their seat.")).toHaveLength(2); // visible, and announced by the status region
    expect(screen.getByText(/all 3 members holding a paid seat are left without one/)).toBeInTheDocument();
    expect(screen.getByText("VAT (10%)")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Pay/ })).toBeEnabled();
  });

  it("names the members who lose a seat and locks paying until it is confirmed", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => jsonResponse(previewFor(1))));
    renderForm();

    fireEvent.change(screen.getByLabelText("Pro seats to renew"), { target: { value: "1" } });

    expect(await screen.findAllByText("2 members lose their seat")).toHaveLength(2); // visible, and announced by the status region
    expect(screen.getByRole("status", { name: "" })).toHaveTextContent("2 members lose their seat");

    const list = screen.getByText("Sara").closest("ul") as HTMLElement;
    expect(within(list).getByText("Sara")).toBeInTheDocument();
    expect(within(list).getByText("Reza")).toBeInTheDocument();
    expect(within(list).getByText("Gets a free seat")).toBeInTheDocument();
    expect(within(list).getByText("Left without a seat")).toBeInTheDocument();
    expect(screen.queryByText("Ali")).not.toBeInTheDocument();

    const pay = screen.getByRole("button", { name: /^Pay/ });
    expect(pay).toBeDisabled();
    expect(screen.getByText("Tick the confirmation above to pay.")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("checkbox", { name: /these 2 members lose their paid seat/ }));
    expect(pay).toBeEnabled();
  });

  it("rejects a count above the seats held", () => {
    renderForm();

    fireEvent.change(screen.getByLabelText("Pro seats to renew"), { target: { value: "9" } });

    expect(screen.getByText("Enter a number from 0 to 3.")).toBeInTheDocument();
  });

  it("asks for nothing before renewal opens: no counts to save, no payment, just the rules", () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    renderForm({ windowOpen: false, preview: null, nowMs: beforeWindow });

    expect(screen.getByText("Renewal is not open yet")).toBeInTheDocument();
    expect(screen.getByText(/Renewal opens on/)).toBeInTheDocument();
    expect(screen.getByText("22 days until the period ends")).toBeInTheDocument();
    expect(screen.getByText("Your paid seats")).toBeInTheDocument();

    expect(screen.queryByLabelText("Pro seats to renew")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Important renewal rules" })).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("states the renewal rules while renewing", () => {
    renderForm();

    expect(screen.getByRole("heading", { name: "Important renewal rules" })).toBeInTheDocument();
    expect(screen.getByText("You choose how many seats to renew at the moment you renew; nothing is saved beforehand.")).toBeInTheDocument();
    expect(screen.getByText("Renewal is not automatic and is open only during the last 7 days of the period.")).toBeInTheDocument();
    expect(screen.getByText(/the members who got their seat later lose it sooner/)).toBeInTheDocument();
    expect(screen.getByText(/gets a free seat only if they have never had one/)).toBeInTheDocument();
    expect(screen.getByText(/the subscription returns to the free plan/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Change members' seats" })).toHaveAttribute("href", "/users");
  });

  it("saves a changed count only when paying", async () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { origin: "http://tenant1.localhost:3020", assign });

    const calls: Array<{ url: string; method?: string; body?: string }> = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) => {
        calls.push({ url, method: init?.method, body: init?.body as string | undefined });
        if (url.includes("/renewal/preview")) return jsonResponse(previewFor(2));
        if (url.endsWith("/renewal")) return jsonResponse(baseSummary);
        return jsonResponse(
          {
            id: "o1", kind: "renewal", seat: "pro", quantity: 2, remaining_days: 31, period_days: 31, payable_irr: 39_600_000,
            status: "PENDING", payment_url: "https://pay.example/gateway", created_at: "2026-10-28T00:00:00Z",
          },
          201,
        );
      }),
    );
    renderForm();

    fireEvent.change(screen.getByLabelText("Pro seats to renew"), { target: { value: "2" } });
    await screen.findAllByText("1 members lose their seat");
    expect(calls.some((call) => call.method === "PUT"), "nothing is saved while choosing").toBe(false);

    fireEvent.click(screen.getByRole("checkbox", { name: /these 1 members lose their paid seat/ }));
    fireEvent.click(screen.getByRole("button", { name: /^Pay/ }));

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://pay.example/gateway"));

    const put = calls.find((call) => call.method === "PUT");
    expect(put?.url).toBe("/api/billing/seats/renewal");
    expect(JSON.parse(put?.body ?? "{}")).toEqual({ seat: "pro", quantity: 2 });
    expect(calls.findIndex((call) => call.method === "PUT")).toBeLessThan(calls.findIndex((call) => call.url === "/api/billing/seats/orders"));
  });

  it("saves the counts, creates the renewal order and goes to the gateway", async () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { origin: "http://tenant1.localhost:3020", assign });

    const calls: Array<{ url: string; method?: string; body?: string }> = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) => {
        calls.push({ url, method: init?.method, body: init?.body as string | undefined });
        if (url.includes("/renewal/preview")) return jsonResponse(previewFor(3));
        if (url.endsWith("/renewal")) return jsonResponse(baseSummary);
        return jsonResponse(
          {
            id: "o1", kind: "renewal", seat: "pro", quantity: 3, remaining_days: 31, period_days: 31, payable_irr: 59_400_000,
            status: "PENDING", payment_url: "https://pay.example/gateway", created_at: "2026-10-28T00:00:00Z",
          },
          201,
        );
      }),
    );
    renderForm();

    fireEvent.click(screen.getByRole("button", { name: /^Pay/ }));

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://pay.example/gateway"));

    const order = calls.find((call) => call.url === "/api/billing/seats/orders");
    expect(JSON.parse(order?.body ?? "{}")).toMatchObject({ kind: "renewal", return_url: "http://tenant1.localhost:3020/billing/seats/renew" });
    expect(calls.some((call) => call.method === "PUT"), "unchanged counts are not saved again").toBe(false);
  });

  it("hides member names from someone who may not list users but still shows what happens", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        const preview = previewFor(2);
        preview.lines[0].losing = preview.lines[0].losing.map((member) => ({ ...member, email: "", name: "" }));
        return jsonResponse(preview);
      }),
    );
    renderForm();

    fireEvent.change(screen.getByLabelText("Pro seats to renew"), { target: { value: "2" } });

    expect(await screen.findByText(/Member \(seeing names needs/)).toBeInTheDocument();
    expect(screen.getByText("Gets a free seat")).toBeInTheDocument();
  });

  it("reports the result of a payment", () => {
    renderForm({ paymentResult: "success" });
    expect(screen.getByText("Renewal done")).toBeInTheDocument();
  });

  it("is read-only without billing management", () => {
    renderForm({ canManage: false });

    expect(screen.getByLabelText("Pro seats to renew")).toBeDisabled();
    expect(screen.queryByRole("button", { name: /^Pay/ })).not.toBeInTheDocument();
    expect(screen.getByText(/needs the billing management permission/)).toBeInTheDocument();
  });
});
