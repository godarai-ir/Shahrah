import { describe, expect, it } from "vitest";

import { extractPromptVariableNames, renderPromptText } from "@/lib/prompt-builder-utils";

describe("prompt management utils", () => {
  it("extracts variable names from prompt text", () => {
    expect(
      extractPromptVariableNames([
        "Hello {{customer_name}}",
        "Order {{order_id}} for {{customer_name}}",
      ]),
    ).toEqual(["customer_name", "order_id"]);
  });

  it("renders template text and reports missing variables", () => {
    const result = renderPromptText("Hello {{customer_name}} from {{city}}", {
      customer_name: "Sara",
    });

    expect(result.rendered).toBe("Hello Sara from {{city}}");
    expect(result.missing).toEqual(["city"]);
  });
});
