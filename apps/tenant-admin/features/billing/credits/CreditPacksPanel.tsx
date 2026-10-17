"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { CreditCard, Landmark, PackageOpen } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Spinner } from "@/components/ui/spinner";
import { useAppLocale, useNumberFormatter } from "@/hooks/use-locale-format";

import { BankTransferPanel } from "../BillingFinancePanels";
import { useBillingFormat } from "../use-billing-format";
import { BANK_TRANSFER_PROVIDER, createCreditPackOrder, type CreditPack, type CreditPackOrder } from "./api";

const ORDER_STATUS_TONE: Record<string, "success" | "warning" | "error" | "pending" | "neutral"> = {
  COMPLETED: "success",
  AWAITING_TRANSFER: "warning",
  PENDING: "pending",
  FAILED: "error",
  CANCELED: "neutral",
};

type PaymentResult = "success" | "failed" | null;

type CreditPacksPanelProps = {
  packs: CreditPack[];
  orders: CreditPackOrder[];
  /** Buying needs billing:manage; without it the packs are shown read-only. */
  canManage: boolean;
  /** The gateway's verdict when the payment page returns here. */
  paymentResult: PaymentResult;
  loadError: string | null;
};

/** Add-on credit packs: buy online or by bank transfer, and follow the orders. */
export function CreditPacksPanel({ packs, orders, canManage, paymentResult, loadError }: CreditPacksPanelProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const locale = useAppLocale();
  const percent = useNumberFormatter({ maximumFractionDigits: 2 });
  const router = useRouter();
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(loadError);
  const [transferOrder, setTransferOrder] = useState<CreditPackOrder | null>(null);

  async function buy(pack: CreditPack, provider: string) {
    setPending(`${pack.slug}:${provider}`);
    setError(null);
    try {
      const order = await createCreditPackOrder({
        pack_slug: pack.slug,
        provider,
        return_url: `${window.location.origin}/billing/credits`,
      });
      if (order.status === "AWAITING_TRANSFER") {
        setTransferOrder(order);
        router.refresh();
      } else if (order.payment_url) {
        window.location.assign(order.payment_url);
        return;
      } else {
        setError(t("credits.packs.noPaymentURL"));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t("unknownError"));
    }
    setPending(null);
  }

  const orderColumns = useMemo<ColumnDef<CreditPackOrder>[]>(
    () => [
      { accessorKey: "credits", header: t("credits.orders.columns.credits"), cell: ({ row }) => <span className="tabular-nums">{format.credits(row.original.credits)}</span> },
      { accessorKey: "payable_irr", header: t("credits.orders.columns.payable"), cell: ({ row }) => <span className="tabular-nums">{format.toman(row.original.payable_irr)}</span> },
      {
        accessorKey: "provider",
        header: t("credits.orders.columns.method"),
        cell: ({ row }) => t(row.original.provider === BANK_TRANSFER_PROVIDER ? "credits.orders.methods.bankTransfer" : "credits.orders.methods.online"),
      },
      {
        accessorKey: "status",
        header: t("credits.orders.columns.status"),
        cell: ({ row }) => (
          <StatusBadge status={ORDER_STATUS_TONE[row.original.status] ?? "neutral"}>
            {t(row.original.status in ORDER_STATUS_TONE ? `credits.orders.statuses.${row.original.status}` : "credits.orders.statuses.other")}
          </StatusBadge>
        ),
      },
      {
        accessorKey: "created_at",
        header: t("credits.orders.columns.createdAt"),
        cell: ({ row }) => <DateTimeValue value={row.original.created_at} locale={locale} />,
      },
      {
        id: "actions",
        header: () => <span className="sr-only">{t("credits.orders.columns.actions")}</span>,
        cell: ({ row }) => row.original.status === "AWAITING_TRANSFER" && canManage ? (
          <Button type="button" size="sm" variant="outline" onClick={() => setTransferOrder(row.original)}>
            {t("credits.orders.submitReceipt")}
          </Button>
        ) : null,
      },
    ],
    [canManage, format, locale, t],
  );

  return (
    <div className="flex flex-col gap-6">
      {paymentResult === "success" ? (
        <Alert>
          <AlertTitle>{t("credits.payment.successTitle")}</AlertTitle>
          <AlertDescription>{t("credits.payment.successDescription")}</AlertDescription>
        </Alert>
      ) : paymentResult === "failed" ? (
        <Alert variant="destructive">
          <AlertTitle>{t("credits.payment.failedTitle")}</AlertTitle>
          <AlertDescription>{t("credits.payment.failedDescription")}</AlertDescription>
        </Alert>
      ) : null}

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {!canManage ? (
        <Alert>
          <AlertDescription>{t("credits.packs.readOnly")}</AlertDescription>
        </Alert>
      ) : null}

      {packs.length === 0 ? (
        <EmptyState icon={PackageOpen} title={t("credits.packs.emptyTitle")} description={t("credits.packs.emptyDescription")} />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {packs.map((pack) => {
            const discount = Number(pack.discount_percent) || 0;
            return (
              <AppCard
                key={`${pack.slug}:${pack.version}`}
                title={<span className="tabular-nums">{format.credits(pack.credits)}</span>}
                description={t("credits.packs.validity", { months: pack.validity_months })}
                actions={discount > 0 ? <StatusBadge status="success">{t("credits.packs.discount", { percent: percent.format(discount) })}</StatusBadge> : undefined}
                footer={
                  <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2">
                    <Button type="button" disabled={!canManage || pending !== null} aria-busy={pending === `${pack.slug}:` || undefined} onClick={() => buy(pack, "")}>
                      {pending === `${pack.slug}:` ? <Spinner data-icon="inline-start" /> : <CreditCard data-icon="inline-start" />}
                      {t("credits.packs.payOnline")}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      disabled={!canManage || pending !== null}
                      aria-busy={pending === `${pack.slug}:${BANK_TRANSFER_PROVIDER}` || undefined}
                      onClick={() => buy(pack, BANK_TRANSFER_PROVIDER)}
                    >
                      {pending === `${pack.slug}:${BANK_TRANSFER_PROVIDER}` ? <Spinner data-icon="inline-start" /> : <Landmark data-icon="inline-start" />}
                      {t("credits.packs.payByTransfer")}
                    </Button>
                  </div>
                }
              >
                <DefinitionList columns={1}>
                  <DefinitionListItem label={t("credits.packs.price")} value={format.toman(pack.price_irr)} />
                  <DefinitionListItem label={t("credits.packs.tax", { percent: percent.format(Number(pack.tax_percent) || 0) })} value={format.toman(pack.tax_irr)} />
                  <DefinitionListItem label={t("credits.packs.payable")} value={format.toman(pack.payable_irr)} emphasis />
                </DefinitionList>
              </AppCard>
            );
          })}
        </div>
      )}

      <AppCard title={t("credits.orders.title")} description={t("credits.orders.description")}>
        <DataTable tableId="billing-credit-orders" columns={orderColumns} data={orders} emptyLabel={t("credits.orders.empty")} />
      </AppCard>

      <FormSheet
        open={transferOrder !== null}
        onOpenChange={(open) => {
          if (!open) setTransferOrder(null);
        }}
        size="lg"
        title={t("credits.transfer.title", { credits: format.credits(transferOrder?.credits ?? 0) })}
        description={t("credits.transfer.description", { amount: format.toman(transferOrder?.payable_irr ?? 0) })}
        closeLabel={t("credits.transfer.close")}
      >
        {transferOrder ? (
          <BankTransferPanel
            key={transferOrder.id}
            initialItems={[]}
            initialError={null}
            packOrder={{ id: transferOrder.id, payableIRR: transferOrder.payable_irr }}
            onSubmitted={() => setTransferOrder(null)}
          />
        ) : null}
      </FormSheet>
    </div>
  );
}
