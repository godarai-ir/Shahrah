import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { SearchPickerDialog } from "./SearchPickerDialog";

describe("SearchPickerDialog", () => {
  it("exposes a controlled search before picker content", () => {
    const onSearchValueChange = vi.fn();
    render(<SearchPickerDialog open onOpenChange={() => undefined} title="Choose model" searchValue="" onSearchValueChange={onSearchValueChange} searchPlaceholder="Search models"><button>GPT</button></SearchPickerDialog>);
    fireEvent.change(screen.getByPlaceholderText("Search models"), { target: { value: "gpt" } });
    expect(onSearchValueChange).toHaveBeenCalledWith("gpt");
    expect(screen.getByRole("button", { name: "GPT" })).toBeInTheDocument();
  });
});
