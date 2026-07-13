import { fireEvent, render, screen } from "@testing-library/react";
import { Shield } from "lucide-react";
import { describe, expect, it, vi } from "vitest";

import { GatewayPolicyRulesList, type GatewayPolicyRulesListProps } from "./GatewayPolicyRulesList";

function createProps(overrides: Partial<GatewayPolicyRulesListProps> = {}): GatewayPolicyRulesListProps {
  return {
    title: "Policy rules",
    createLabel: "New rule",
    onCreate: vi.fn(),
    alertLabels: { error: "Error", success: "Success" },
    query: "",
    onQueryChange: vi.fn(),
    searchPlaceholder: "Search rules",
    resultCountLabel: "0 results",
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
  it("composes the list toolbar, search, and empty-state create action without a duplicate page heading", () => {
    const onCreate = vi.fn();
    const onQueryChange = vi.fn();
    render(<GatewayPolicyRulesList {...createProps({ onCreate, onQueryChange })} />);

    expect(screen.queryByRole("heading", { name: "Policy rules" })).not.toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Policy rules" })).toBeInTheDocument();
    expect(screen.getByText("No rules")).toBeInTheDocument();
    fireEvent.change(screen.getByRole("searchbox", { name: "Search rules" }), { target: { value: "monthly" } });
    fireEvent.click(screen.getAllByRole("button", { name: "New rule" })[0]);
    expect(onQueryChange).toHaveBeenCalledWith("monthly");
    expect(onCreate).toHaveBeenCalledOnce();
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
