"use client";

import { ArrowDownLeft, ReceiptText, TicketPercent, WalletCards } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import type { DiscountRedemption, Invoice, WalletTransaction } from "./domain";
import { getCurrentJalaliDateValue, normalizeJalaliDateInput } from "./domain";
import { cn } from "@/lib/utils";

export type BillingProfileType = "individual" | "legal";

/** Translator bound to the `billing` namespace. */
export type BillingTranslator = (key: string, values?: Record<string, string | number>) => string;

export function ProfileTypeField({
  value,
  onValueChange,
}: {
  value: BillingProfileType;
  onValueChange: (value: BillingProfileType) => void;
}) {
  const t = useTranslations("billing");
  const options = [
    { value: "individual" as const, label: t("profileType.individual") },
    { value: "legal" as const, label: t("profileType.legal") },
  ];

  return (
    <div className="flex flex-col gap-2">
      <Label>{t("profileType.label")}</Label>
      <RadioGroup
        value={value}
        onValueChange={(nextValue) => onValueChange(nextValue as BillingProfileType)}
        className="grid grid-cols-2 gap-2"
      >
        {options.map((option) => (
          <Label
            key={option.value}
            htmlFor={`profile-type-${option.value}`}
            className={cn(
              "flex cursor-pointer items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm font-medium transition-colors rtl:flex-row",
              value === option.value && "border-primary bg-accent/40 text-accent-foreground",
            )}
          >
            <span>{option.label}</span>
            <RadioGroupItem id={`profile-type-${option.value}`} value={option.value} />
          </Label>
        ))}
      </RadioGroup>
    </div>
  );
}

export function matchInvoiceFilters(item: Invoice, status: string) {
  return status === "all" || item.status === status;
}

export function matchTransactionFilters(item: WalletTransaction, status: string, type: string) {
  if (status !== "all" && item.status !== status) return false;
  return type === "all" || item.type === type;
}

export function matchDiscountFilters(item: DiscountRedemption, status: string, type: string) {
  if (status !== "all" && item.status !== status) return false;
  return type === "all" || item.discount_type === type;
}

export function buildFilterOptions(
  values: string[],
  labeler: (value: string) => string,
  allLabel: string,
) {
  const uniqueValues = Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
  return [
    { value: "all", label: allLabel },
    ...uniqueValues.map((value) => ({ value, label: labeler(value) })),
  ];
}

const DOCUMENT_STATUSES = new Set([
  "successful",
  "completed",
  "approved",
  "paid",
  "used",
  "active",
  "issued",
  "pending",
  "pending_review",
  "draft",
  "failed",
  "rejected",
  "cancelled",
  "expired",
  "refunded",
]);
const TRANSACTION_TYPES = new Set(["consumption", "topup", "refund", "adjustment"]);
const DISCOUNT_TYPES = new Set(["fixed", "percentage", "gift_credit"]);

/** Falls back to the raw control-plane value when the code is not yet mapped. */
export function mapDocumentStatusLabel(status: string, t: BillingTranslator) {
  return DOCUMENT_STATUSES.has(status) ? t(`status.${status}`) : status || "-";
}

export function mapTransactionTypeLabel(type: string, t: BillingTranslator) {
  return TRANSACTION_TYPES.has(type) ? t(`transactionType.${type}`) : type || "-";
}

export function mapDiscountTypeLabel(type: string, t: BillingTranslator) {
  return DISCOUNT_TYPES.has(type) ? t(`discountType.${type}`) : type || "-";
}

export function transactionTypeIcon(type: string) {
  switch (type) {
    case "consumption":
      return <ArrowDownLeft className="size-4 text-destructive" aria-hidden="true" />;
    case "topup":
      return <WalletCards className="size-4 text-primary" aria-hidden="true" />;
    case "refund":
      return <ReceiptText className="size-4 text-warning" aria-hidden="true" />;
    case "adjustment":
      return <TicketPercent className="size-4 text-muted-foreground" aria-hidden="true" />;
    default:
      return <ReceiptText className="size-4 text-muted-foreground" aria-hidden="true" />;
  }
}

export function downloadCSV(headers: string[], rows: string[][], filename: string) {
  const csvLines = [headers.join(","), ...rows.map((row) => row.map(escapeCSVCell).join(","))];
  const blob = new Blob(["\uFEFF" + csvLines.join("\n")], { type: "text/csv;charset=utf-8;" });
  const objectURL = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = objectURL;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(objectURL);
}

function escapeCSVCell(value: string) {
  const normalized = String(value ?? "");
  if (normalized.includes(",") || normalized.includes("\"") || normalized.includes("\n")) {
    return `"${normalized.replaceAll("\"", "\"\"")}"`;
  }
  return normalized;
}

export function normalizeDigits(value: string) {
  return value
    .trim()
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)))
    .replace(/\D/g, "");
}

export function JalaliDateField({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  const t = useTranslations("billing");

  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="flex gap-2">
        <Input
          id={id}
          dir="ltr"
          inputMode="numeric"
          value={value}
          onChange={(event) => onChange(normalizeJalaliDateInput(event.target.value))}
          placeholder="1405-04-17"
        />
        <Button type="button" variant="outline" onClick={() => onChange(getCurrentJalaliDateValue())}>
          {t("jalaliDate.today")}
        </Button>
      </div>
      <p className="text-xs leading-5 text-muted-foreground">{t("jalaliDate.hint")}</p>
    </div>
  );
}

export function formatFileSize(size: number) {
  if (size >= 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  if (size >= 1024) return `${Math.round(size / 1024)} KB`;
  return `${size} B`;
}
