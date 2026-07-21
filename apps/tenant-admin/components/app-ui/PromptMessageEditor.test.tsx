import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PromptMessageEditor } from "./PromptMessageEditor";

const roleOptions = [
  { value: "system" as const, label: "System" },
  { value: "user" as const, label: "User" },
  { value: "assistant" as const, label: "Assistant" },
  { value: "tool" as const, label: "Tool" },
];

function renderEditor(error?: string) {
  render(
    <PromptMessageEditor
      role="user"
      content="Hello"
      roleOptions={roleOptions}
      roleLabel="Message role"
      contentLabel="Message content"
      toolCallIdLabel="Tool call ID"
      duplicateLabel="Duplicate"
      removeLabel="Remove"
      dragLabel="Reorder message"
      moveUpLabel="Move up"
      moveDownLabel="Move down"
      error={error}
      onRoleChange={vi.fn()}
      onContentChange={vi.fn()}
      onDuplicate={vi.fn()}
      onRemove={vi.fn()}
      onMoveUp={vi.fn()}
      onMoveDown={vi.fn()}
    />,
  );
}

describe("PromptMessageEditor", () => {
  it("gives the role picker its own accessible name", () => {
    renderEditor();

    expect(screen.getByRole("combobox", { name: "Message role" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Message content" })).toBeInTheDocument();
  });

  it("connects the content error to the textarea", () => {
    renderEditor("Message is required");

    const textarea = screen.getByRole("textbox", { name: "Message content" });
    const error = screen.getByText("Message is required");
    expect(textarea).toHaveAttribute("aria-invalid", "true");
    expect(textarea).toHaveAttribute("aria-describedby", error.id);
  });
});
