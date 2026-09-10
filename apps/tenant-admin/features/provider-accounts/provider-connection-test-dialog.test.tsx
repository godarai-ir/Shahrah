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

  it("shows the TokenBazaar billable warning and disables Run until the spend is explicitly approved", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "tokenbazaar" })} />,
    );

    expect(screen.getByText("providerAccounts.ux.connectionTest.tokenbazaarBillableWarning")).toBeInTheDocument();

    const runButton = screen.getByRole("button", { name: /connectionTest.run/ });
    expect(runButton).toBeDisabled();

    const checkbox = screen.getByText(
      "providerAccounts.ux.connectionTest.tokenbazaarApprovalCheckbox",
    ).closest("label")?.querySelector("input[type=checkbox]");
    expect(checkbox).toBeTruthy();

    fireEvent.click(checkbox as Element);

    expect(runButton).not.toBeDisabled();
  });

  it("never fires the billable test automatically: it only runs after an explicit Run click", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "tokenbazaar" })} />,
    );

    expect(testProviderAccountCredentials).not.toHaveBeenCalled();

    const checkbox = screen.getByText(
      "providerAccounts.ux.connectionTest.tokenbazaarApprovalCheckbox",
    ).closest("label")?.querySelector("input[type=checkbox]");
    fireEvent.click(checkbox as Element);

    expect(testProviderAccountCredentials).not.toHaveBeenCalled();

    fireEvent.submit(document.getElementById("provider-connection-test-form") as HTMLFormElement);

    expect(testProviderAccountCredentials).toHaveBeenCalledTimes(1);
    expect(testProviderAccountCredentials).toHaveBeenCalledWith("acct-1", "");
  });

  it("blocks submission (even via form submit) while TokenBazaar spend is not approved", () => {
    render(
      <ProviderConnectionTestDialog open onOpenChange={() => {}} account={account({ provider: "tokenbazaar" })} />,
    );

    fireEvent.submit(document.getElementById("provider-connection-test-form") as HTMLFormElement);

    expect(testProviderAccountCredentials).not.toHaveBeenCalled();
  });
});
