import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PrincipalPicker } from "./principal-picker";

vi.mock("@/app/components/principal-search-input", () => ({
  PrincipalEntitySelect: (props: { ariaLabel?: string; onSelectedChange: (value: unknown) => void }) => (
    <button
      type="button"
      aria-label={props.ariaLabel}
      onClick={() => props.onSelectedChange({ type: "everyone", display_name: "Everyone" })}
    >
      Select everyone
    </button>
  ),
}));

const labels = {
  users: "Users",
  teams: "Teams",
  orgUnits: "Org units",
  virtualAccounts: "Virtual accounts",
  externalIdentities: "External identities",
  everyone: "Everyone",
  searchPlaceholder: "Search",
  triggerPlaceholder: "Search",
  loading: "Loading",
  empty: "Empty",
};

describe("PrincipalPicker", () => {
  it("requires explicit confirmation before granting access to everyone", () => {
    const setSelected = vi.fn();
    render(
      <PrincipalPicker
        id="consumers"
        title="Consumers"
        searchLabel="Search consumers"
        onSearchPrincipal={async () => []}
        principalSelectorLabels={labels}
        selected={[]}
        setSelected={setSelected}
        allowEveryone
        addLabel="Add consumer"
        removeLabel="Remove consumer"
        everyoneConfirmation={{
          title: "Grant access to everyone?",
          description: "This is broad access.",
          confirmLabel: "Confirm broad access",
          cancelLabel: "Cancel",
        }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Search consumers" }));
    fireEvent.click(screen.getByRole("button", { name: "Add consumer" }));
    expect(screen.getByRole("dialog", { name: "Grant access to everyone?" })).toBeInTheDocument();
    expect(setSelected).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Confirm broad access" }));
    expect(setSelected).toHaveBeenCalledWith([{ type: "everyone", principal_id: undefined, display_name: "Everyone" }]);
  });

  it("uses a principal-specific accessible name for removal", () => {
    render(
      <PrincipalPicker
        id="managers"
        title="Managers"
        onSearchPrincipal={async () => []}
        principalSelectorLabels={labels}
        selected={[{ type: "user", principal_id: "user-1", display_name: "Ada" }]}
        setSelected={() => {}}
        allowEveryone={false}
        addLabel="Add manager"
        removeLabel="Remove manager"
        getRemoveLabel={(item) => `Remove ${item.display_name} from managers`}
      />,
    );

    expect(screen.getByRole("button", { name: "Remove Ada from managers" })).toBeInTheDocument();
  });
});
