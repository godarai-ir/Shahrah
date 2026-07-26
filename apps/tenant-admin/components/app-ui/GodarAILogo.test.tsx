import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { GodarAILogo } from "./GodarAILogo";

describe("GodarAILogo", () => {
  it("renders light and dark sources for the automatic variant without changing the intrinsic ratio", () => {
    const { container } = render(
      <GodarAILogo decorative className="h-10 w-auto max-w-full object-contain" />,
    );
    const images = [...container.querySelectorAll("img")];

    expect(images).toHaveLength(2);
    expect(images[0]).toHaveAttribute("src", expect.stringContaining("godarai-mark-light.svg"));
    expect(images[1]).toHaveAttribute("src", expect.stringContaining("godarai-mark-dark.svg"));
    for (const image of images) {
      expect(image).toHaveAttribute("width", "336");
      expect(image).toHaveAttribute("height", "247");
      expect(image).toHaveAttribute("alt", "");
      expect(image).toHaveClass("h-10", "w-auto", "max-w-full", "object-contain");
    }
  });

  it("renders one accessible source for an explicit variant", () => {
    render(<GodarAILogo variant="mono-navy" alt="GodarAI logo" />);

    const image = screen.getByAltText("GodarAI logo");
    expect(image).toHaveAttribute("src", expect.stringContaining("godarai-mark-mono-navy.svg"));
  });
});
