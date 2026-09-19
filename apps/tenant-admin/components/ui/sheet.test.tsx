import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Sheet, SheetContent, SheetTitle } from "./sheet";

describe("SheetContent side", () => {
  it("defaults to the inline-end edge so it opens on the reading-end side in both RTL and LTR", () => {
    render(
      <Sheet open>
        <SheetContent>
          <SheetTitle>Panel</SheetTitle>
        </SheetContent>
      </Sheet>,
    );
    const dialog = screen.getByRole("dialog");
    expect(dialog.className).toMatch(/\bend-0\b/);
    expect(dialog.className).toMatch(/\bborder-s\b/);
    expect(dialog.className).not.toMatch(/\bstart-0\b/);
  });

  it("uses the inline-start edge when explicitly requested", () => {
    render(
      <Sheet open>
        <SheetContent side="inline-start">
          <SheetTitle>Panel</SheetTitle>
        </SheetContent>
      </Sheet>,
    );
    const dialog = screen.getByRole("dialog");
    expect(dialog.className).toMatch(/\bstart-0\b/);
    expect(dialog.className).toMatch(/\bborder-e\b/);
    expect(dialog.className).not.toMatch(/\bend-0\b/);
  });

  it("keeps top and bottom sheets unaffected by the logical-axis change", () => {
    render(
      <Sheet open>
        <SheetContent side="bottom">
          <SheetTitle>Panel</SheetTitle>
        </SheetContent>
      </Sheet>,
    );
    expect(screen.getByRole("dialog").className).toMatch(/\bbottom-0\b/);
  });
});
