import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { CodePanel } from "./CodePanel";

describe("CodePanel", () => {
  it("renders its heading, language, code, and copy action", () => {
    render(<CodePanel title="Request" description="Gateway request" displayLanguage="cURL" code="curl /v1/models" language="bash" copy={{ copyLabel: "Copy code", copiedLabel: "Copied" }} />);
    expect(screen.getByText("Request")).toBeInTheDocument();
    expect(screen.getByText("cURL")).toBeInTheDocument();
    expect(screen.getByText(/curl/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Copy code" })).toBeInTheDocument();
  });
});
