"use client";

import { useCallback, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { ShoppingCart } from "lucide-react";

import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAppLocale } from "@/hooks/use-locale-format";
import { APIClientError } from "@/lib/http/api-client";
import { buySeatsHref, seatErrorMessage } from "@/lib/seats";

import type { CreditSettings } from "../credits/api";
import { useBillingFormat } from "../use-billing-format";
import {
  approveSeatRequest,
  denySeatRequest,
  purchaseRequiredQuote,
  type SeatOrder,
  type SeatRequest,
  type SeatSummary,
} from "./api";
import { SeatRenewalCard } from "./SeatRenewalCard";
import { SeatRequestsTable } from "./SeatRequestsTable";
import { SeatSettingsCard } from "./SeatSettingsCard";
import { SeatSummaryCards } from "./SeatSummaryCards";

const ORDER_TONE: Record<string, "success" | "pending" | "error" | "neutral"> = {
  COMPLETED: "success",
  PENDING: "pending",
  FAILED: "error",
  CANCELED: "neutral",
};

type SeatsWorkspaceProps = {
  summary: SeatSummary;
  requests: SeatRequest[];
  orders: SeatOrder[];
  settings: CreditSettings | null;
  /** Assigning seats needs user:manage. */
  canManageMembers: boolean;
  /** Buying and renewing seats needs billing:manage. */
  canManageBilling: boolean;
  paymentResult: "success" | "failed" | null;
};

/** Seat stock: buying and renewing paid seats, upgrade requests, and orders. Members are managed on the users page. */
export function SeatsWorkspace(props: SeatsWorkspaceProps) {
  const { requests, orders, settings, canManageMembers, canManageBilling, paymentResult } = props;
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const locale = useAppLocale();
  const router = useRouter();
  const [summary, setSummary] = useState(props.summary);
  const [pendingID, setPendingID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const openRequests = requests.filter((request) => request.status === "pending").length;

  const refresh = useCallback(() => router.refresh(), [router]);

  const approve = useCallback(
    async (request: SeatRequest) => {
      setPendingID(request.id);
      setError(null);
      try {
        await approveSeatRequest(request.id);
        showFeedbackToast({ kind: "success", title: t("requests.approved") });
        refresh();
      } catch (err) {
        if (err instanceof APIClientError && purchaseRequiredQuote(err.payload) !== undefined) {
          router.push(buySeatsHref(request.requested_seat, request.user_id, request.id));
        } else {
          setError(seatErrorMessage(t, err));
        }
      }
      setPendingID(null);
    },
    [refresh, router, t],
  );

  const deny = useCallback(
    async (request: SeatRequest, note: string) => {
      try {
        await denySeatRequest(request.id, note);
        refresh();
      } catch (err) {
        setError(seatErrorMessage(t, err));
      }
    },
    [refresh, t],
  );

  const orderColumns = useMemo<ColumnDef<SeatOrder>[]>(
    () => [
      { id: "kind", header: t("orders.columns.kind"), cell: ({ row }) => t(`orders.kinds.${row.original.kind}`) },
      {
        id: "seats",
        header: t("orders.columns.seats"),
        cell: ({ row }) => t("orders.seatsValue", { count: row.original.quantity, seat: t(`seats.${row.original.seat}`) }),
      },
      { accessorKey: "payable_irr", header: t("orders.columns.payable"), cell: ({ row }) => <span className="tabular-nums">{format.toman(row.original.payable_irr)}</span> },
      {
        accessorKey: "status",
        header: t("orders.columns.status"),
        cell: ({ row }) => (
          <StatusBadge status={ORDER_TONE[row.original.status] ?? "neutral"}>
            {t(row.original.status in ORDER_TONE ? `orders.statuses.${row.original.status}` : "orders.statuses.other")}
          </StatusBadge>
        ),
      },
      { accessorKey: "created_at", header: t("orders.columns.createdAt"), cell: ({ row }) => <DateTimeValue value={row.original.created_at} locale={locale} /> },
    ],
    [format, locale, t],
  );

  return (
    <div className="flex flex-col gap-6">
      {paymentResult === "success" ? (
        <Alert>
          <AlertTitle>{t("payment.successTitle")}</AlertTitle>
          <AlertDescription>{t("payment.successDescription")}</AlertDescription>
        </Alert>
      ) : paymentResult === "failed" ? (
        <Alert variant="destructive">
          <AlertTitle>{t("payment.failedTitle")}</AlertTitle>
          <AlertDescription>{t("payment.failedDescription")}</AlertDescription>
        </Alert>
      ) : null}

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errors.title")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {/* The primary action leads the page; the stock it changes follows. */}
      <div className="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
        <p className="max-w-3xl text-sm leading-6 text-muted-foreground">{t("intro")}</p>
        {canManageBilling ? (
          <Button asChild>
            <Link href="/billing/seats/buy">
              <ShoppingCart data-icon="inline-start" />
              {t("buy.open")}
            </Link>
          </Button>
        ) : null}
      </div>

      <SeatSummaryCards summary={summary} />


      <Tabs defaultValue="requests">
        <TabsList>
          <TabsTrigger value="requests">
            {t("tabs.requests")}
            {openRequests > 0 ? <Badge variant="secondary" className="ms-2 tabular-nums">{format.count(openRequests)}</Badge> : null}
          </TabsTrigger>
          <TabsTrigger value="orders">{t("tabs.orders")}</TabsTrigger>
        </TabsList>
        <TabsContent value="requests" className="pt-4">
          <SeatRequestsTable requests={requests} canManage={canManageMembers} pendingID={pendingID} onApprove={approve} onDeny={deny} />
        </TabsContent>
        <TabsContent value="orders" className="pt-4">
          <DataTable tableId="seat-orders" columns={orderColumns} data={orders} getRowId={(order) => order.id} emptyLabel={t("orders.empty")} />
        </TabsContent>
      </Tabs>

      <SeatRenewalCard summary={summary} canManage={canManageBilling} onSummaryChange={setSummary} />
      {settings ? <SeatSettingsCard initialSettings={settings} canManage={canManageBilling} /> : null}
    </div>
  );
}
