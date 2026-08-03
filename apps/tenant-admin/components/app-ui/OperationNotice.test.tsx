import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { OperationNotice, type OperationNoticeLabels } from "./OperationNotice";

const labels: OperationNoticeLabels = { success: "Success", error: "Error", warning: "Warning", info: "Information" };

describe("OperationNotice", () => {
  it("renders nothing without a notice", () => {
    const { container } = render(<OperationNotice labels={labels} />);
    expect(container).toBeEmptyDOMElement();
  });

  it.each([
    ["success", "Success", "status"],
    ["error", "Error", "alert"],
    ["warning", "Warning", "status"],
    ["info", "Information", "status"],
  ] as const)("maps %s feedback to its localized title", (kind, title, role) => {
    render(<OperationNotice notice={{ kind, message: `${kind} message` }} labels={labels} />);
    expect(screen.getByRole(role)).toHaveTextContent(title);
    expect(screen.getByRole(role)).toHaveTextContent(`${kind} message`);
  });
});
