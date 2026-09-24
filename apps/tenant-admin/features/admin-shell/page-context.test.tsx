import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import {
  AdminShellPageBridge,
  AdminShellPageProvider,
  useAdminShellPage,
} from "./page-context";

describe("AdminShell page metadata bridge", () => {
  it("publishes route metadata and resets it when the route unmounts", async () => {
    const { rerender } = render(
      <AdminShellPageProvider>
        <PageStateProbe />
        <AdminShellPageBridge
          pageHeader={{ title: "Logs", description: "Workspace logs", helpHref: "/docs/logs" }}
          contentPadding="none"
          contentLayout="fill"
        >
          <div>Page content</div>
        </AdminShellPageBridge>
      </AdminShellPageProvider>,
    );

    expect(await screen.findByTestId("page-title")).toHaveTextContent("Logs");
    expect(screen.getByTestId("page-description")).toHaveTextContent("Workspace logs");
    expect(screen.getByTestId("page-help")).toHaveTextContent("/docs/logs");
    expect(screen.getByTestId("page-padding")).toHaveTextContent("none");
    expect(screen.getByTestId("page-layout")).toHaveTextContent("fill");

    rerender(
      <AdminShellPageProvider>
        <PageStateProbe />
      </AdminShellPageProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("page-title")).toBeEmptyDOMElement();
      expect(screen.getByTestId("page-padding")).toHaveTextContent("default");
      expect(screen.getByTestId("page-layout")).toHaveTextContent("flow");
    });
  });
});

function PageStateProbe() {
  const { pageHeader, contentPadding, contentLayout } = useAdminShellPage();

  return (
    <div>
      <span data-testid="page-title">{pageHeader?.title}</span>
      <span data-testid="page-description">{pageHeader?.description}</span>
      <span data-testid="page-help">{pageHeader?.helpHref}</span>
      <span data-testid="page-padding">{contentPadding}</span>
      <span data-testid="page-layout">{contentLayout}</span>
    </div>
  );
}
