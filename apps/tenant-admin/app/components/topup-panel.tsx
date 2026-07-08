"use client";

import { FormEvent, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";

import { AppSelectField, SettingsCard, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type WalletTopUpSettings = {
  min_amount_usd: number;
  max_amount_usd: number;
  tax_percent: number;
  fee_percent: number;
};

type Wallet = {
  id: string;
  tenant_id: string;
  balance_irr: number;
  pending_irr: number;
};

type WalletTransaction = {
  id: string;
  type: string;
  amount_irr: number;
  status: string;
  created_at: string;
};

type TopUpIntentResponse = {
  payment_url?: string;
};

type Props = {
  initialWallet: Wallet | null;
  initialSettings: WalletTopUpSettings | null;
  initialTransactions: WalletTransaction[];
  initialError: string | null;
};

export default function TopUpPanel(props: Props) {
  const t = useTranslations("topupPage");
  const common = useTranslations("common");

  const [amountToman, setAmountToman] = useState("1000000");
  const [provider, setProvider] = useState("mock");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(props.initialError);

  const formattedBalance = useMemo(() => {
    if (!props.initialWallet) return common("none");
    return formatIRR(props.initialWallet.balance_irr);
  }, [common, props.initialWallet]);

  const columns = useMemo<ColumnDef<WalletTransaction>[]>(
    () => [
      {
        accessorKey: "id",
        header: t("transactions.columns.id"),
        cell: ({ row }) => <span className="font-mono text-xs">{row.original.id || "-"}</span>,
      },
      {
        accessorKey: "type",
        header: t("transactions.columns.type"),
      },
      {
        accessorKey: "status",
        header: t("transactions.columns.status"),
        cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge>,
      },
      {
        accessorKey: "amount_irr",
        header: t("transactions.columns.amount"),
        cell: ({ row }) => formatIRR(row.original.amount_irr),
      },
      {
        accessorKey: "created_at",
        header: t("transactions.columns.createdAt"),
        cell: ({ row }) => new Date(row.original.created_at).toLocaleString("fa-IR"),
      },
    ],
    [t],
  );

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const toman = Number(amountToman.replaceAll(",", ""));
    if (!Number.isFinite(toman) || toman <= 0) {
      setError(t("messages.amountInvalid"));
      return;
    }

    const amountIRR = Math.round(toman * 10);
    setLoading(true);
    try {
      const returnURL = `${window.location.origin}/billing/topup/result`;
      const resp = await fetch("/api/wallet/topup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          amount_irr: amountIRR,
          provider,
          return_url: returnURL,
          metadata: { source: "tenant-admin-topup-page" },
        }),
      });

      const payload = (await resp.json()) as TopUpIntentResponse & { error?: { message?: string } };
      if (!resp.ok) {
        throw new Error(payload.error?.message ?? t("messages.intentFailed"));
      }
      if (!payload.payment_url) {
        throw new Error(t("messages.missingRedirect"));
      }
      window.location.assign(payload.payment_url);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("messages.unknownError"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <SettingsCard title={t("title")} description={t("description")}>
          <form onSubmit={onSubmit} className="grid gap-4 md:grid-cols-3">
            <div className="grid gap-2">
              <Label htmlFor="amountToman">{t("amountLabel")}</Label>
              <Input
                id="amountToman"
                inputMode="numeric"
                value={amountToman}
                onChange={(e) => setAmountToman(e.target.value)}
              />
            </div>

            <AppSelectField
              label={t("providerLabel")}
              value={provider}
              onValueChange={setProvider}
              options={[
                { value: "mock", label: t("providers.mock") },
                { value: "zarinpal", label: t("providers.zarinpal") },
                { value: "idpay", label: t("providers.idpay") },
              ]}
            />

            <div className="grid gap-2">
              <Label>{t("balanceLabel")}</Label>
              <div className="flex h-10 items-center rounded-md border px-3 text-sm">{formattedBalance}</div>
            </div>

            <div className="md:col-span-3 text-xs text-muted-foreground">
              {t("constraints", {
                tax: String(props.initialSettings?.tax_percent ?? 0),
                fee: String(props.initialSettings?.fee_percent ?? 2),
              })}
            </div>

            <div className="md:col-span-3">
              <Button type="submit" disabled={loading}>
                {loading ? t("submitLoading") : t("submit")}
              </Button>
            </div>
          </form>
      </SettingsCard>

      <SettingsCard title={t("transactions.title")} description={t("transactions.description")}>
        <DataTable columns={columns} data={props.initialTransactions} emptyLabel={t("transactions.empty")} />
      </SettingsCard>
    </div>
  );
}

const numberFormatter = new Intl.NumberFormat("fa-IR");

function formatIRR(value: number): string {
  return `${numberFormatter.format(Math.round(value))} IRR`;
}

function statusTone(status: string): "success" | "warning" | "error" | "pending" | "neutral" {
  if (["successful", "completed", "approved", "paid", "used"].includes(status)) return "success";
  if (["pending", "pending_review", "draft"].includes(status)) return "pending";
  if (["failed", "rejected", "cancelled", "expired"].includes(status)) return "error";
  if (status === "refunded") return "warning";
  return "neutral";
}
