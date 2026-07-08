import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { RateLimitRuleForm } from "./RateLimitRuleForm";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn(), refresh: vi.fn() }),
}));

vi.mock("./ruleApi", () => ({
  createRateLimitRule: vi.fn(async () => ({})),
  updateRateLimitRule: vi.fn(async () => ({})),
}));

describe("RateLimitRuleForm", () => {
  it("adds and removes condition rows", async () => {
    render(<RateLimitRuleForm />);
    fireEvent.click(screen.getByText("actions.addFilter"));
    expect(screen.getAllByLabelText("حذف فیلتر")).toHaveLength(1);
    fireEvent.click(screen.getByLabelText("حذف فیلتر"));
    expect(screen.queryByLabelText("حذف فیلتر")).toBeNull();
  });

  it("save disabled on invalid form", () => {
    render(<RateLimitRuleForm />);
    const submit = screen.getByRole("button", { name: "actions.save" });
    expect(submit).toBeDisabled();
  });

  it("advanced options can expand and collapse", () => {
    render(<RateLimitRuleForm />);
    const toggle = screen.getByText("نمایش تنظیمات پیشرفته");
    fireEvent.click(toggle);
    expect(screen.getByText("رفتار تعارض")).toBeInTheDocument();
    fireEvent.click(screen.getByText("بستن تنظیمات پیشرفته"));
    expect(screen.queryByText("رفتار تعارض")).toBeNull();
  });
});
