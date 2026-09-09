import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { RateLimitRuleForm } from "./RateLimitRuleForm";

const translations: Record<string, string> = {
  "actions.add": "Add",
  "actions.addFilter": "Add Filter",
  "actions.save": "Save",
  "form.advanced.toggleOpen": "Show advanced settings",
  "form.advanced.toggleClose": "Hide advanced settings",
  "form.fields.conflictBehavior": "Conflict behavior",
  "aria.removeFilter": "Remove filter",
  "form.switches.enabled.title": "Rule Status",
  "form.switches.softLimit.title": "Soft Limit",
  "form.switches.retryAfterHeader.title": "Send Retry-After Header",
};

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
  useTranslations: () => (key: string, values?: Record<string, string | number>) => {
    const template = translations[key] ?? key;
    if (!values) return template;
    return Object.entries(values).reduce((result, [name, value]) => result.replace(`{${name}}`, String(value)), template);
  },
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
    fireEvent.click(screen.getByText("Add Filter"));
    expect(screen.getByRole("button", { name: /^Add$/ })).toBeDisabled();
    expect(screen.getAllByLabelText("Remove filter")).toHaveLength(1);
    fireEvent.click(screen.getByLabelText("Remove filter"));
    expect(screen.queryByLabelText("Remove filter")).toBeNull();
  });

  it("save disabled on invalid form", () => {
    render(<RateLimitRuleForm />);
    const submit = screen.getByRole("button", { name: "Save" });
    expect(submit).toBeDisabled();
  });

  it("advanced options can expand and collapse", () => {
    render(<RateLimitRuleForm />);
    const toggle = screen.getByText("Show advanced settings");
    fireEvent.click(toggle);
    expect(screen.getByRole("group", { name: "Conflict behavior" })).toBeInTheDocument();
    fireEvent.click(screen.getByText("Hide advanced settings"));
    expect(screen.queryByRole("group", { name: "Conflict behavior" })).toBeNull();
  });

  it("switches expose accessible names", () => {
    render(<RateLimitRuleForm />);
    expect(screen.getByRole("switch", { name: "Rule Status" })).toBeInTheDocument();
    fireEvent.click(screen.getByText("Show advanced settings"));
    expect(screen.getByRole("switch", { name: "Soft Limit" })).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: "Send Retry-After Header" })).toBeInTheDocument();
  });
});
