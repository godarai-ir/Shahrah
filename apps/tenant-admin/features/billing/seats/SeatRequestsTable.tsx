"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";

import { AppTextareaField } from "@/components/app-ui/AppTextField";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Spinner } from "@/components/ui/spinner";
import { useAppLocale } from "@/hooks/use-locale-format";

import type { SeatRequest } from "./api";

const REQUEST_TONE: Record<SeatRequest["status"], "pending" | "success" | "error" | "neutral"> = {
  pending: "pending",
  approved: "success",
  denied: "error",
  canceled: "neutral",
};

type SeatRequestsTableProps = {
  requests: SeatRequest[];
  canManage: boolean;
  pendingID: string | null;
  onApprove: (request: SeatRequest) => void;
  onDeny: (request: SeatRequest, note: string) => Promise<void>;
};

/** Members asking for a better seat, newest pending first. */
export function SeatRequestsTable({ requests, canManage, pendingID, onApprove, onDeny }: SeatRequestsTableProps) {
  const t = useTranslations("billingSeats");
  const locale = useAppLocale();
  const [denying, setDenying] = useState<SeatRequest | null>(null);
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const columns = useMemo<ColumnDef<SeatRequest>[]>(
    () => [
      {
        id: "member",
        header: t("requests.columns.member"),
        cell: ({ row }) => <span className="font-medium">{row.original.name || row.original.email}</span>,
      },
      {
        id: "change",
        header: t("requests.columns.change"),
        cell: ({ row }) =>
          t("requests.changeValue", {
            from: t(`seats.${["free", "pro", "max"].includes(row.original.current_seat) ? row.original.current_seat : "unassigned"}`),
            to: t(`seats.${row.original.requested_seat}`),
          }),
      },
      {
        id: "reason",
        header: t("requests.columns.reason"),
        cell: ({ row }) => (
          <div className="flex max-w-72 flex-col gap-1">
            {row.original.reason ? <span className="line-clamp-2">{row.original.reason}</span> : <span className="text-muted-foreground">—</span>}
            {row.original.block_reason ? <span className="text-xs text-muted-foreground">{t(`blockReasons.${row.original.block_reason}`)}</span> : null}
          </div>
        ),
      },
      {
        accessorKey: "status",
        header: t("requests.columns.status"),
        cell: ({ row }) => <StatusBadge status={REQUEST_TONE[row.original.status]}>{t(`requests.statuses.${row.original.status}`)}</StatusBadge>,
      },
      {
        accessorKey: "created_at",
        header: t("requests.columns.createdAt"),
        cell: ({ row }) => <DateTimeValue value={row.original.created_at} locale={locale} />,
      },
      {
        id: "actions",
        header: () => <span className="sr-only">{t("requests.columns.actions")}</span>,
        cell: ({ row }) =>
          row.original.status === "pending" && canManage ? (
            <div className="flex gap-2">
              <Button type="button" size="sm" disabled={pendingID !== null} aria-busy={pendingID === row.original.id || undefined} onClick={() => onApprove(row.original)}>
                {pendingID === row.original.id ? <Spinner data-icon="inline-start" /> : null}
                {t("requests.approve")}
              </Button>
              <Button type="button" size="sm" variant="outline" disabled={pendingID !== null} onClick={() => { setNote(""); setDenying(row.original); }}>
                {t("requests.deny")}
              </Button>
            </div>
          ) : null,
      },
    ],
    [canManage, locale, onApprove, pendingID, t],
  );

  async function submitDeny() {
    if (!denying) return;
    setSubmitting(true);
    try {
      await onDeny(denying, note);
      setDenying(null);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <DataTable tableId="seat-requests" columns={columns} data={requests} getRowId={(request) => request.id} emptyLabel={t("requests.empty")} />
      <FormDialog
        open={denying !== null}
        onOpenChange={(open) => !open && setDenying(null)}
        pending={submitting}
        title={t("requests.denyTitle", { member: denying ? denying.name || denying.email : "" })}
        footer={
          <Button type="button" variant="destructive" disabled={submitting} aria-busy={submitting || undefined} onClick={submitDeny}>
            {submitting ? <Spinner data-icon="inline-start" /> : null}
            {t("requests.denyConfirm")}
          </Button>
        }
      >
        <AppTextareaField label={t("requests.denyNote")} value={note} maxLength={1000} onChange={(event) => setNote(event.target.value)} />
      </FormDialog>
    </>
  );
}
