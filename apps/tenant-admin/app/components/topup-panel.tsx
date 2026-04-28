"use client";

import { FormEvent, useMemo, useState } from "react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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

  const [amountUSD, setAmountUSD] = useState("100");
  const [provider, setProvider] = useState("mock");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(props.initialError);

  const minUSD = props.initialSettings?.min_amount_usd ?? 1;
  const maxUSD = props.initialSettings?.max_amount_usd ?? 1000;

  const formattedBalance = useMemo(() => {
    if (!props.initialWallet) return common("none");
    return formatIRR(props.initialWallet.balance_irr);
  }, [common, props.initialWallet]);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const amount = Number(amountUSD);
    if (!Number.isFinite(amount) || amount < minUSD || amount > maxUSD) {
      setError(t("messages.amountOutOfRange", { min: String(minUSD), max: String(maxUSD) }));
      return;
    }

    setLoading(true);
    try {
      const returnURL = `${window.location.origin}/billing/topup/result`;
      const resp = await fetch("/api/wallet/topup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          amount_usd: amount,
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

      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="grid gap-4 md:grid-cols-3">
            <div className="grid gap-2">
              <Label htmlFor="amountUSD">{t("amountLabel")}</Label>
              <Input id="amountUSD" value={amountUSD} onChange={(e) => setAmountUSD(e.target.value)} />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="provider">{t("providerLabel")}</Label>
              <select
                id="provider"
                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                value={provider}
                onChange={(e) => setProvider(e.target.value)}
              >
                <option value="mock">{t("providers.mock")}</option>
                <option value="zarinpal">{t("providers.zarinpal")}</option>
                <option value="idpay">{t("providers.idpay")}</option>
              </select>
            </div>

            <div className="grid gap-2">
              <Label>{t("balanceLabel")}</Label>
              <div className="flex h-10 items-center rounded-md border px-3 text-sm">{formattedBalance}</div>
            </div>

            <div className="md:col-span-3 text-xs text-muted-foreground">
              {t("constraints", {
                min: String(minUSD),
                max: String(maxUSD),
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
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("transactions.title")}</CardTitle>
          <CardDescription>{t("transactions.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("transactions.columns.id")}</TableHead>
                <TableHead>{t("transactions.columns.type")}</TableHead>
                <TableHead>{t("transactions.columns.status")}</TableHead>
                <TableHead>{t("transactions.columns.amount")}</TableHead>
                <TableHead>{t("transactions.columns.createdAt")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {props.initialTransactions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground">
                    {t("transactions.empty")}
                  </TableCell>
                </TableRow>
              ) : (
                props.initialTransactions.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="font-mono text-xs">{item.id}</TableCell>
                    <TableCell>{item.type}</TableCell>
                    <TableCell>{item.status}</TableCell>
                    <TableCell>{formatIRR(item.amount_irr)}</TableCell>
                    <TableCell>{new Date(item.created_at).toLocaleString("fa-IR")}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}

const numberFormatter = new Intl.NumberFormat("fa-IR");

function formatIRR(value: number): string {
  return `${numberFormatter.format(Math.round(value))} IRR`;
}
