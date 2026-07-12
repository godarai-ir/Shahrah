import { fireEvent, render, screen } from "@testing-library/react";
import { Database, Shield } from "lucide-react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { StatusOptionGroup, type StatusOption } from "./StatusOptionGroup";

const options = [
  {
    value: "metadata",
    badge: "metadata-only",
    title: "Metadata only",
    description: "Store metadata without content.",
    icon: Database,
    tone: "info",
  },
  {
    value: "full",
    badge: "full-content",
    title: "Full content",
    description: "Store complete content.",
    icon: Shield,
    tone: "success",
  },
] satisfies readonly StatusOption<"metadata" | "full">[];

function ControlledGroup({ onValueChange = vi.fn() }: { onValueChange?: (value: "metadata" | "full") => void }) {
  const [value, setValue] = useState<"metadata" | "full">("metadata");

  return (
    <StatusOptionGroup
      label="Capture mode"
      value={value}
      options={options}
      onValueChange={(nextValue) => {
        setValue(nextValue);
        onValueChange(nextValue);
      }}
    />
  );
}

describe("StatusOptionGroup", () => {
  it("exposes a labeled single-selection group", () => {
    render(<ControlledGroup />);

    expect(screen.getByRole("group", { name: "Capture mode" })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Metadata only" })).toHaveAttribute("data-state", "on");
    expect(screen.getByRole("radio", { name: "Full content" })).toHaveAttribute("data-state", "off");
  });

  it("changes the selected option and does not clear it when selected again", () => {
    const onValueChange = vi.fn();
    render(<ControlledGroup onValueChange={onValueChange} />);

    const fullContent = screen.getByRole("radio", { name: "Full content" });
    fireEvent.click(fullContent);
    expect(fullContent).toHaveAttribute("data-state", "on");
    expect(onValueChange).toHaveBeenLastCalledWith("full");

    fireEvent.click(fullContent);
    expect(fullContent).toHaveAttribute("data-state", "on");
    expect(onValueChange).toHaveBeenCalledTimes(1);
  });

  it("disables all options while the feature is saving", () => {
    render(
      <StatusOptionGroup
        label="Capture mode"
        value="metadata"
        options={options}
        onValueChange={vi.fn()}
        disabled
      />,
    );

    for (const option of screen.getAllByRole("radio")) {
      expect(option).toBeDisabled();
    }
  });
});
