import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CopyAction } from "./CopyAction";

describe("CopyAction", () => {
  const writeText = vi.fn();

  beforeEach(() => {
    writeText.mockReset();
    writeText.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText } });
  });

  it("copies its value and exposes the copied state", async () => {
    render(<CopyAction value="resource-id" copyLabel="Copy ID" copiedLabel="ID copied" />);
    fireEvent.click(screen.getByRole("button", { name: "Copy ID" }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith("resource-id"));
    expect(screen.getByRole("button", { name: "ID copied" })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("ID copied");
  });

  it("reports copy failures to the owner", async () => {
    const onCopyError = vi.fn();
    writeText.mockRejectedValueOnce(new Error("denied"));
    render(<CopyAction value="resource-id" copyLabel="Copy ID" copiedLabel="ID copied" onCopyError={onCopyError} />);
    fireEvent.click(screen.getByRole("button", { name: "Copy ID" }));
    await waitFor(() => expect(onCopyError).toHaveBeenCalledOnce());
    expect(screen.getByRole("button", { name: "Copy ID" })).toBeInTheDocument();
  });
});
