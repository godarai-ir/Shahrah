import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { WIFCredentialStatus } from "./wif-credential-status";

vi.mock("next-intl", () => ({
  useLocale: () => "en-US",
  useTranslations: () => (key: string) => key,
}));

describe("WIFCredentialStatus", () => {
  it("renders sanitized scopes, expiry, and refresh time", () => {
    render(
      <WIFCredentialStatus
        result={{
          success: true,
          wif_effective_scopes: [
            "workspace:developer",
            "workspace:inference",
          ],
          wif_expires_at: "2026-07-29T12:00:00Z",
          wif_last_refresh_at: "2026-07-29T11:00:00Z",
        }}
      />,
    );

    expect(
      screen.getByText("workspace:developer workspace:inference"),
    ).toBeInTheDocument();
    expect(screen.getByText("wifExpiresAt")).toBeInTheDocument();
    expect(screen.getByText("wifLastRefreshAt")).toBeInTheDocument();
    expect(document.querySelectorAll("time")).toHaveLength(2);
  });

  it("does not render for a non-WIF probe result", () => {
    const { container } = render(
      <WIFCredentialStatus result={{ success: true }} />,
    );
    expect(container).toBeEmptyDOMElement();
  });
});
