import { render, screen } from "@testing-library/react";
import { Shield } from "lucide-react";
import { describe, expect, it, vi } from "vitest";

import { GatewayPolicyRulesList, type GatewayPolicyRulesListProps } from "./GatewayPolicyRulesList";

function createProps(overrides: Partial<GatewayPolicyRulesListProps> = {}): GatewayPolicyRulesListProps {
  return {
    title: "Policy rules",
    createAction: <button type="button">New rule</button>,
    alertLabels: { error: "Error", success: "Success" },
    toolbar: <input type="search" aria-label="Search rules" />,
    filtered: false,
    itemCount: 0,
    loading: false,
    loadingLabel: "Loading rules",
    emptyIcon: Shield,
    emptyTitle: "No rules",
    emptyDescription: "Create the first rule.",
    children: <div>Rules table</div>,
    ruleSheet: {
      open: false,
      onOpenChange: vi.fn(),
      title: "Create rule",
      description: "Configure the rule.",
      closeLabel: "Close",
      content: <div>Rule form</div>,
    },
    deleteDialog: {
      open: false,
      onOpenChange: vi.fn(),
      title: "Delete rule",
      description: "This cannot be undone.",
      cancelLabel: "Cancel",
      confirmLabel: "Delete",
      pendingLabel: "Deleting",
      pending: false,
      onConfirm: vi.fn(),
    },
    upgradeDialog: {
      open: false,
      onOpenChange: vi.fn(),
      title: "Upgrade",
      description: "Upgrade the plan.",
      cancelLabel: "Cancel",
      actionLabel: "Upgrade plan",
    },
    ...overrides,
  };
}

describe("GatewayPolicyRulesList", () => {
  it("composes the toolbar, aside, and empty-state create action without a duplicate page heading", () => {
    render(<GatewayPolicyRulesList {...createProps({ aside: <div>Simulator</div> })} />);

    expect(screen.queryByRole("heading", { name: "Policy rules" })).not.toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Policy rules" })).toBeInTheDocument();
    expect(screen.getByText("No rules")).toBeInTheDocument();
    expect(screen.getByRole("searchbox", { name: "Search rules" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "New rule" })).toHaveLength(2);
    expect(screen.getByRole("complementary")).toHaveTextContent("Simulator");
  });

  it("drops the empty-state create action while the list is filtered", () => {
    render(<GatewayPolicyRulesList {...createProps({ filtered: true })} />);
    expect(screen.getAllByRole("button", { name: "New rule" })).toHaveLength(1);
  });

  it("keeps the feature form inside an accessible controlled sheet", () => {
    render(
      <GatewayPolicyRulesList
        {...createProps({
          ruleSheet: {
            open: true,
            onOpenChange: vi.fn(),
            title: "Create rule",
            description: "Configure the rule.",
            closeLabel: "Close",
            content: <div>Rule form</div>,
          },
        })}
      />,
    );

    expect(screen.getByRole("dialog", { name: "Create rule" })).toBeInTheDocument();
    expect(screen.getByText("Rule form")).toBeInTheDocument();
  });
});
