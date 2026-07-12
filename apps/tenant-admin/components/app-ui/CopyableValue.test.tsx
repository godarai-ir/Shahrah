import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CopyableValue } from "./CopyableValue";

const labels = {
  valueAriaLabel: "API token",
  copy: "Copy token",
  copying: "Copying",
  copied: "Copied",
  reveal: "Reveal token",
  hide: "Hide token",
};

describe("CopyableValue", () => {
  const writeText = vi.fn();

  beforeEach(() => {
    writeText.mockReset();
    writeText.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });
  });

  it("copies the raw value and announces success", async () => {
    render(<CopyableValue value="secret-value" labels={labels} />);

    fireEvent.click(screen.getByRole("button", { name: "Copy token" }));

    await waitFor(() => expect(writeText).toHaveBeenCalledWith("secret-value"));
    expect(await screen.findByRole("status")).toHaveTextContent("Copied");
    expect(screen.getByRole("button", { name: "Copied" })).toBeInTheDocument();
  });

  it("supports a masked value with an explicit reveal action", () => {
    render(<CopyableValue value="secret-value" labels={labels} masked allowReveal />);

    const input = screen.getByLabelText("API token");
    expect(input).toHaveAttribute("type", "password");
    fireEvent.click(screen.getByRole("button", { name: "Reveal token" }));
    expect(input).toHaveAttribute("type", "text");
  });
});
