import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { LocaleToggle, LocaleToggleMenuItem } from "./locale-toggle";
import { DropdownMenu, DropdownMenuContent } from "./ui/dropdown-menu";

const refresh = vi.fn();
let activeLocale = "fa";

const TRANSLATIONS: Record<string, string> = {
  switch: "Change language",
  "names.fa": "فارسی",
  "names.en": "English",
};

vi.mock("next-intl", () => ({
  useLocale: () => activeLocale,
  useTranslations: () => (key: string) => TRANSLATIONS[key] ?? key,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

/** Radix opens on key events; jsdom has no real pointer events. */
function openToggle() {
  render(<LocaleToggle />);
  fireEvent.keyDown(screen.getByRole("button", { name: "Change language" }), { key: "Enter" });
}

function renderMenuItem() {
  return render(
    <DropdownMenu open>
      <DropdownMenuContent>
        <LocaleToggleMenuItem />
      </DropdownMenuContent>
    </DropdownMenu>,
  );
}

describe("LocaleToggleMenuItem", () => {
  beforeEach(() => {
    refresh.mockReset();
    activeLocale = "fa";
    document.cookie = "NEXT_LOCALE=; Path=/; Max-Age=0";
    document.documentElement.lang = "fa";
    document.documentElement.dir = "rtl";
  });

  afterEach(() => {
    document.cookie = "NEXT_LOCALE=; Path=/; Max-Age=0";
  });

  it("offers the other locale while Persian is active", () => {
    renderMenuItem();

    expect(screen.getByRole("menuitem")).toHaveTextContent("English");
  });

  it("offers Persian while English is active", () => {
    activeLocale = "en";
    renderMenuItem();

    expect(screen.getByRole("menuitem")).toHaveTextContent("فارسی");
  });

  it("stores the choice and flips the document direction to LTR", () => {
    renderMenuItem();
    fireEvent.click(screen.getByRole("menuitem"));

    expect(document.cookie).toContain("NEXT_LOCALE=en");
    expect(document.documentElement.lang).toBe("en");
    expect(document.documentElement.dir).toBe("ltr");
    expect(refresh).toHaveBeenCalledOnce();
  });

  it("restores RTL when switching back to Persian", () => {
    activeLocale = "en";
    renderMenuItem();
    fireEvent.click(screen.getByRole("menuitem"));

    expect(document.cookie).toContain("NEXT_LOCALE=fa");
    expect(document.documentElement.dir).toBe("rtl");
  });
});

describe("LocaleToggle", () => {
  beforeEach(() => {
    refresh.mockReset();
    activeLocale = "fa";
    document.cookie = "NEXT_LOCALE=; Path=/; Max-Age=0";
    document.documentElement.dir = "rtl";
  });

  it("lists both languages and switches to the one the reader picks", () => {
    openToggle();

    const options = screen.getAllByRole("menuitemradio");
    expect(options).toHaveLength(2);

    fireEvent.click(screen.getByRole("menuitemradio", { name: "English" }));

    expect(document.cookie).toContain("NEXT_LOCALE=en");
    expect(document.documentElement.dir).toBe("ltr");
    expect(refresh).toHaveBeenCalledOnce();
  });

  it("does nothing when the active language is picked again", () => {
    openToggle();
    fireEvent.click(screen.getByRole("menuitemradio", { name: "فارسی" }));

    expect(refresh).not.toHaveBeenCalled();
  });
});
