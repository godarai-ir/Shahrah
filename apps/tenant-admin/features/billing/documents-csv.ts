import type { DiscountRedemption, Invoice, WalletTransaction } from "@/features/billing/domain";
import {
  downloadCSV,
  mapDiscountTypeLabel,
  mapDocumentStatusLabel,
  mapTransactionTypeLabel,
  type BillingTranslator,
} from "@/features/billing/billing-ui-helpers";
import type { BillingFormat } from "@/features/billing/use-billing-format";

export type BillingDocumentTab = "invoices" | "transactions" | "discounts";

type ExportInput = {
  tab: BillingDocumentTab;
  invoices: Invoice[];
  transactions: WalletTransaction[];
  discounts: DiscountRedemption[];
  t: BillingTranslator;
  format: BillingFormat;
};

/** Exports the rows currently visible in the given documents tab. */
export function downloadBillingDocumentsCSV({ tab, invoices, transactions, discounts, t, format }: ExportInput) {
  if (tab === "invoices") {
    downloadCSV(
      ["number", "status", "periodEnd", "finalAmount", "netAdjustment"].map((key) => t(`documents.invoices.${key}`)),
      invoices.map((item) => [
        item.invoice_number,
        mapDocumentStatusLabel(item.status, t),
        format.dateTime(item.period_end),
        format.amount(item.final_amount_irr),
        format.amount(item.net_adjustment_irr),
      ]),
      "billing-invoices.csv",
    );
    return;
  }

  if (tab === "transactions") {
    downloadCSV(
      ["type", "status", "amount", "balanceAfter", "createdAt"].map((key) => t(`documents.transactions.${key}`)),
      transactions.map((item) => [
        mapTransactionTypeLabel(item.type, t),
        mapDocumentStatusLabel(item.status, t),
        format.amount(item.amount_irr),
        format.amount(item.balance_after_irr),
        format.dateTime(item.created_at),
      ]),
      "billing-transactions.csv",
    );
    return;
  }

  downloadCSV(
    ["code", "type", "status", "usedAmount", "createdAt"].map((key) => t(`documents.discounts.${key}`)),
    discounts.map((item) => [
      item.code,
      mapDiscountTypeLabel(item.discount_type, t),
      mapDocumentStatusLabel(item.status, t),
      format.amount(item.used_amount_irr),
      format.dateTime(item.created_at),
    ]),
    "billing-discounts.csv",
  );
}
