import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "./field";
import { Input } from "./input";

describe("Field", () => {
  it("composes an accessible labelled control with help and error text", () => {
    render(
      <FieldGroup>
        <Field data-invalid>
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <Input id="email" aria-invalid aria-describedby="email-description email-error" />
          <FieldDescription id="email-description">Use your work email.</FieldDescription>
          <FieldError id="email-error">Enter a valid email.</FieldError>
        </Field>
      </FieldGroup>,
    );

    expect(screen.getByRole("textbox", { name: "Email" })).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByText("Use your work email.")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("Enter a valid email.");
  });

  it("deduplicates validator messages", () => {
    render(<FieldError errors={[{ message: "Required" }, { message: "Required" }, { message: "Invalid" }]} />);

    expect(screen.getAllByRole("listitem")).toHaveLength(2);
    expect(screen.getByText("Required")).toBeInTheDocument();
    expect(screen.getByText("Invalid")).toBeInTheDocument();
  });
});
