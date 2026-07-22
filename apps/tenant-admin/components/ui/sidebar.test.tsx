import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { Sidebar, SidebarProvider, SidebarTrigger } from "./sidebar";

const originalInnerWidth = window.innerWidth;

describe("Sidebar", () => {
  beforeEach(() => {
    Object.defineProperty(window, "innerWidth", { configurable: true, value: 500 });
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));
  });

  afterEach(() => {
    Object.defineProperty(window, "innerWidth", { configurable: true, value: originalInnerWidth });
    vi.unstubAllGlobals();
  });

  it("localizes and preserves RTL semantics for the mobile sheet", async () => {
    render(
      <SidebarProvider>
        <SidebarTrigger aria-label="باز کردن منو" />
        <Sidebar
          side="right"
          dir="rtl"
          mobileTitle="ناوبری پنل"
          mobileDescription="دسترسی سریع به ابزارهای پنل"
          mobileCloseLabel="بستن منو"
        >
          <nav aria-label="مسیریابی اصلی">محتوای منو</nav>
        </Sidebar>
      </SidebarProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "باز کردن منو" }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveAttribute("dir", "rtl");
    expect(screen.getByRole("heading", { name: "ناوبری پنل" })).toBeInTheDocument();
    expect(screen.getByText("دسترسی سریع به ابزارهای پنل")).toBeInTheDocument();
    const closeButton = screen.getByRole("button", { name: "بستن منو" });
    expect(closeButton).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "مسیریابی اصلی" })).toBeInTheDocument();

    fireEvent.click(closeButton);
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
