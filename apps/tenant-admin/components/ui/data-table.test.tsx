import type { ColumnDef } from "@tanstack/react-table";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { DataTableBulkBar } from "@/components/app-ui/DataTableBulkBar";
import { DataTableToolbar } from "@/components/app-ui/DataTableToolbar";

import { DataTable } from "./data-table";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string, values?: Record<string, unknown>) =>
    values ? `${key}:${JSON.stringify(values)}` : key,
}));

type Row = { id: string; name: string; usage: number };

const rows: Row[] = [
  { id: "u1", name: "Ada", usage: 12 },
  { id: "u2", name: "Linus", usage: 7 },
];

const columns: ColumnDef<Row>[] = [
  { id: "name", header: "Name", cell: ({ row }) => row.original.name },
  { id: "usage", header: "Usage", meta: { optional: true, label: "Usage" }, cell: ({ row }) => `usage-${row.original.usage}` },
];

function openMenu(name: string) {
  const trigger = screen.getByRole("button", { name });
  fireEvent.pointerDown(trigger, { button: 0, ctrlKey: false, pointerType: "mouse" });
}

describe("DataTable", () => {
  beforeEach(() => window.localStorage.clear());

  it("hides an optional column from the column menu", async () => {
    render(<DataTable tableId="t" columns={columns} data={rows} emptyLabel="empty" toolbar={<DataTableToolbar />} />);
    expect(screen.getByText("usage-12")).toBeInTheDocument();

    openMenu("table.columns");
    fireEvent.click(await screen.findByRole("menuitemcheckbox", { name: "Usage" }));

    expect(screen.queryByText("usage-12")).not.toBeInTheDocument();
    expect(screen.queryByRole("columnheader", { name: "Usage" })).not.toBeInTheDocument();
    expect(JSON.parse(window.localStorage.getItem("table-prefs:t")!)).toMatchObject({ hiddenColumns: ["usage"] });
  });

  it("starts default-hidden optional columns hidden", () => {
    const hidden: ColumnDef<Row>[] = [columns[0], { ...columns[1], meta: { optional: true, defaultHidden: true, label: "Usage" } }];
    render(<DataTable tableId="t" columns={hidden} data={rows} emptyLabel="empty" />);
    expect(screen.queryByText("usage-12")).not.toBeInTheDocument();
  });

  it("reports selected ids and swaps the toolbar for the bulk bar", () => {
    const onSelectionChange = vi.fn();
    render(
      <DataTable
        tableId="t"
        columns={columns}
        data={rows}
        emptyLabel="empty"
        enableRowSelection
        getRowId={(row) => row.id}
        onSelectionChange={onSelectionChange}
        toolbar={<DataTableToolbar search={<input aria-label="search" />} />}
        bulkBar={(ids, clear) => <DataTableBulkBar count={ids.length} onClear={clear} actions={[]} />}
      />,
    );

    const secondRow = screen.getByText("Linus").closest("tr")!;
    fireEvent.click(within(secondRow).getByRole("checkbox", { name: "table.selectRow" }));

    expect(onSelectionChange).toHaveBeenLastCalledWith(["u2"]);
    expect(secondRow).toHaveAttribute("data-state", "selected");
    expect(screen.queryByRole("textbox", { name: "search" })).not.toBeInTheDocument();
    const bulkBar = screen.getByRole("toolbar", { name: "table.bulkActions" });
    expect(bulkBar).toHaveTextContent('table.selectedCount:{"count":1}');

    fireEvent.click(within(bulkBar).getByRole("button", { name: "table.clearSelection" }));
    expect(onSelectionChange).toHaveBeenLastCalledWith([]);
    expect(screen.getByRole("textbox", { name: "search" })).toBeInTheDocument();
  });

  it("renders skeleton rows while loading", () => {
    render(<DataTable tableId="t" columns={columns} data={rows} emptyLabel="empty" state="loading" skeletonRows={3} />);
    expect(screen.getByText("loading")).toBeInTheDocument();
    expect(screen.queryByText("Ada")).not.toBeInTheDocument();
    expect(screen.getAllByRole("row", { hidden: true })).toHaveLength(4);
  });

  it("renders the no-results slot for an empty page", () => {
    render(<DataTable tableId="t" columns={columns} data={[]} emptyLabel="empty" noResultsSlot={<p>no matches</p>} />);
    expect(screen.getByText("no matches")).toBeInTheDocument();
    expect(screen.queryByText("empty")).not.toBeInTheDocument();
  });
});
