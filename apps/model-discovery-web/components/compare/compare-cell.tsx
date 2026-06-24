import { cn } from "@/lib/utils";

const EMPTY = "-";

export function CompareCell({
  value,
  isDifferent,
}: {
  value: string | number | null | undefined;
  isDifferent: boolean;
}) {
  return (
    <span className={cn("block", isDifferent ? "font-medium text-zinc-100" : "text-zinc-500")}>{value ?? EMPTY}</span>
  );
}

export function compareScalarValues(values: Array<string | number | null | undefined>): boolean {
  const normalized = values.map((value) => (value === null || value === undefined || value === "" ? EMPTY : String(value)));
  return new Set(normalized).size > 1;
}

export function compareListValues(values: string[][]): boolean {
  const normalized = values.map((list) => list.map((v) => v.trim()).filter(Boolean).sort().join("|"));
  return new Set(normalized).size > 1;
}
