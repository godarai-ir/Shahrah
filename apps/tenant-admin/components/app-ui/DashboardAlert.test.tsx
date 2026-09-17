import { render, screen } from "@testing-library/react";
import { AlertTriangle } from "lucide-react";
import { describe, expect, it } from "vitest";

import { DashboardAlertList, type DashboardAlertItem } from "./DashboardAlert";

const item = (id: string, tone: DashboardAlertItem["tone"]): DashboardAlertItem => ({
  id,
  tone,
  icon: AlertTriangle,
  title: id,
  action: { label: `open ${id}`, href: `/${id}` },
});

describe("DashboardAlertList", () => {
  it("orders by severity and collapses overflow into a link", () => {
    render(
      <DashboardAlertList
        items={[item("info", "info"), item("warn", "warning"), item("err", "error"), item("info2", "info")]}
        moreLabel={(count) => `+${count} more`}
      />,
    );

    const tones = [...document.querySelectorAll("[data-tone]")].map((node) => node.getAttribute("data-tone"));
    expect(tones).toEqual(["error", "warning", "info"]);
    expect(screen.getByRole("link", { name: "+1 more" })).toHaveAttribute("href", "/info2");
  });
});
