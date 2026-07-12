import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { TableBody } from "@/components/ui/table";
import { TableEmptyRow, TableLoadingRows } from "./TableStateRows";

describe("table state rows", () => {
  it("renders the requested loading rows with an accessible label", () => {
    const { container } = render(<table><TableBody><TableLoadingRows colSpan={3} rowCount={2} label="Loading users" /></TableBody></table>);
    expect(container.querySelectorAll("tr")).toHaveLength(2);
    expect(screen.getByText("Loading users")).toBeInTheDocument();
  });

  it("renders an empty state spanning all columns", () => {
    render(<table><TableBody><TableEmptyRow colSpan={4} title="No users" description="Change the filters." /></TableBody></table>);
    expect(screen.getByRole("cell")).toHaveAttribute("colspan", "4");
    expect(screen.getByText("No users")).toBeInTheDocument();
  });
});
