import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Sparkline } from "./Sparkline";

describe("Sparkline", () => {
  it("renders nothing for empty values", () => {
    const { container } = render(<Sparkline values={[]} stroke="var(--chart-2)" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders one bar per value", () => {
    const { container } = render(<Sparkline values={[1, 0, 3]} stroke="var(--chart-3)" variant="bars" />);
    expect(container.querySelectorAll("rect")).toHaveLength(3);
  });
});
