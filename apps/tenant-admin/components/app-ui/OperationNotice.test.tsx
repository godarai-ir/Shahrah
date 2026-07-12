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
    ["success", "Success"], ["error", "Error"], ["warning", "Warning"], ["info", "Information"],
  ] as const)("maps %s feedback to its localized title", (kind, title) => {
    render(<OperationNotice notice={{ kind, message: `${kind} message` }} labels={labels} />);
    expect(screen.getByRole("alert")).toHaveTextContent(title);
    expect(screen.getByRole("alert")).toHaveTextContent(`${kind} message`);
  });
});
