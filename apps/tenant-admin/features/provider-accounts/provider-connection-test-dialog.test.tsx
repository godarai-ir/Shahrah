import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ProviderConnectionTestDialog } from "./provider-connection-test-dialog";
import type { TenantProviderAccount } from "./types";

vi.mock("next-intl", () => ({
  useLocale: () => "en-US",
  useTranslations: (namespace: string) => (key: string, values?: Record<string, unknown>) => {
    if (values) return `${namespace}.${key}:${JSON.stringify(values)}`;
    return `${namespace}.${key}`;
  },
}));

const testProviderAccountCredentials = vi.fn();
vi.mock("./api", () => ({
  testProviderAccountCredentials: (...args: unknown[]) => testProviderAccountCredentials(...args),
}));

function account(overrides: Partial<TenantProviderAccount> = {}): TenantProviderAccount {
  return {
    id: "acct-1",
    provider: "openai",
    name: "Primary",
    auth_mode: "api_key",
    has_credential: true,
    use_platform_managed_credentials: false,
    ...overrides,
  } as TenantProviderAccount;
}

describe("ProviderConnectionTestDialog", () => {
  beforeEach(() => {
    testProviderAccountCredentials.mockClear();
  });

  it("does not gate non-TokenBazaar providers behind an approval checkbox", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "openai" })} />,
    );

    expect(screen.queryByText("providerAccounts.ux.connectionTest.tokenbazaarApprovalCheckbox")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /connectionTest.run/ })).not.toBeDisabled();
  });

  it("does not gate TokenBazaar behind an approval checkbox (its credential probe is a free GET /v1/models)", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "tokenbazaar" })} />,
    );

    expect(screen.queryByText("providerAccounts.ux.connectionTest.tokenbazaarBillableWarning")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /connectionTest.run/ })).not.toBeDisabled();
  });

  it("only fires the test after an explicit Run click", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "tokenbazaar" })} />,
    );

    expect(testProviderAccountCredentials).not.toHaveBeenCalled();

    fireEvent.submit(document.getElementById("provider-connection-test-form") as HTMLFormElement);

    expect(testProviderAccountCredentials).toHaveBeenCalledTimes(1);
    expect(testProviderAccountCredentials).toHaveBeenCalledWith("acct-1", "");
  });
});
