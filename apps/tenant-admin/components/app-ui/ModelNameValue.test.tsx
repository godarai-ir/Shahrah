import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ModelNameValue } from "./ModelNameValue";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("ModelNameValue", () => {
  it("renders LTR, left-aligned model text and copies the full value", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText } });

    render(<ModelNameValue value="openai:gpt-5-nano +2" copyValue="openai:gpt-5-nano,openai:gpt-5-mini" />);

    const model = screen.getByText("openai:gpt-5-nano +2");
    expect(model.parentElement).toHaveAttribute("dir", "ltr");
    expect(model.parentElement).toHaveClass("text-left");
    expect(model).toHaveClass("whitespace-nowrap");

    fireEvent.click(screen.getByRole("button", { name: "modelName.copy" }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith("openai:gpt-5-nano,openai:gpt-5-mini"));
  });
});
