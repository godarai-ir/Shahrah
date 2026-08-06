"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useLocale, useTranslations } from "next-intl";
import { BadgePercent, CreditCard, Landmark, TestTubeDiagonal, WalletCards } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { SettingsCard } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Skeleton } from "@/components/ui/skeleton";
import { convertTomanToUSD, formatApproximateUSD, parsePositiveToman } from "@/lib/topup-currency";
import { cn } from "@/lib/utils";

type TopUpIntentResponse = {
  payment_url?: string;
};

type CurrentExchangeRate = {
  base: "USD";
  quote: "IRR";
  rate: number;
  source: string;
  fetchedAt: string;
  status: "fresh" | "stale";
};

type Props = {
  initialError: string | null;
};

type ProviderOption = {
  value: string;
  label: string;
  description: string;
  icon: LucideIcon;
  logoSrc?: string;
};

export default function TopUpPanel(props: Props) {
  const t = useTranslations("topupPage");
  const locale = useLocale();

  const [amountToman, setAmountToman] = useState("");
  const [provider, setProvider] = useState("mock");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(props.initialError);
  const [exchangeRate, setExchangeRate] = useState<CurrentExchangeRate | null>(null);
  const [exchangeRateLoading, setExchangeRateLoading] = useState(true);

  const quickAmounts = [
    { value: "1000000", label: t("quickAmounts.oneMillion") },
    { value: "3000000", label: t("quickAmounts.threeMillion") },
    { value: "10000000", label: t("quickAmounts.tenMillion") },
    { value: "50000000", label: t("quickAmounts.fiftyMillion") },
  ];

  const providerOptions: readonly ProviderOption[] = [
    { value: "mock", label: t("providers.mock"), description: t("providerDescriptions.mock"), icon: TestTubeDiagonal },
    {
      value: "zarinpal",
      label: t("providers.zarinpal"),
      description: t("providerDescriptions.zarinpal"),
      icon: Landmark,
      logoSrc: "/images/providers/zarinpal.png",
    },
    {
      value: "idpay",
      label: t("providers.idpay"),
      description: t("providerDescriptions.idpay"),
      icon: CreditCard,
      logoSrc: "/images/providers/idpay.png",
    },
  ];

  useEffect(() => {
    const controller = new AbortController();

    void (async () => {
      try {
        const response = await fetch("/api/exchange-rates/current", {
          cache: "no-store",
          signal: controller.signal,
        });
        if (!response.ok) throw new Error("exchange rate unavailable");

        const payload = await response.json() as unknown;
        const parsed = parseCurrentExchangeRate(payload);
        if (!parsed) throw new Error("invalid exchange rate response");

        setExchangeRate(parsed);
      } catch {
        if (!controller.signal.aborted) setExchangeRate(null);
      } finally {
        if (!controller.signal.aborted) setExchangeRateLoading(false);
      }
    })();

    return () => controller.abort();
  }, []);

  const parsedToman = parsePositiveToman(amountToman);
  const estimatedUSD = parsedToman !== null && exchangeRate
    ? convertTomanToUSD(parsedToman, exchangeRate.rate)
    : null;
  const approximateUSD = estimatedUSD === null ? null : formatApproximateUSD(estimatedUSD);
  const rateToman = exchangeRate
    ? new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(exchangeRate.rate / 10)
    : null;
  const syncedAt = exchangeRate
    ? new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" }).format(new Date(exchangeRate.fetchedAt))
    : null;

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const toman = parsePositiveToman(amountToman);
    if (toman === null) {
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
    <div className="flex flex-col gap-6">
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <Alert variant="warning">
        <AlertTitle>{t("gatewayNotice.title")}</AlertTitle>
        <AlertDescription>
          {t.rich("gatewayNotice.body", {
            bankTransferLink: (chunks) => (
              <Link href="/billing/bank-transfer" className="font-medium text-warning underline-offset-4 hover:underline">
                {chunks}
              </Link>
            ),
          })}
        </AlertDescription>
      </Alert>

      <SettingsCard
        title={
          <span className="inline-flex items-center gap-2">
            <WalletCards className="size-5 text-primary" />
            {t("title")}
          </span>
        }
      >
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <Field>
            <FieldLabel htmlFor="amountToman">{t("amountLabel")}</FieldLabel>
            <Input
              id="amountToman"
              dir="ltr"
              inputMode="numeric"
              value={amountToman}
              onChange={(e) => setAmountToman(e.target.value)}
            />
            <div className="flex flex-wrap gap-2">
              {quickAmounts.map((amount) => (
                <Button
                  key={amount.value}
                  type="button"
                  variant={amountToman === amount.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => setAmountToman(amount.value)}
                >
                  <WalletCards data-icon="inline-start" />
                  {amount.label}
                </Button>
              ))}
            </div>

            <div
              className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3"
              role="status"
              aria-live="polite"
            >
              {exchangeRateLoading ? (
                <>
                  <span className="sr-only">{t("exchangeRate.loading")}</span>
                  <Skeleton className="h-5 w-48 max-w-full" />
                  <Skeleton className="h-4 w-64 max-w-full" />
                </>
              ) : exchangeRate && rateToman && syncedAt ? (
                <>
                  <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                    <div className="flex flex-col gap-1">
                      <span className="text-sm font-medium text-foreground">{t("exchangeRate.liveLabel")}</span>
                      <span className="text-sm font-semibold tabular-nums text-foreground">
                        {t("exchangeRate.rateValue", { rate: rateToman })}
                      </span>
                    </div>
                    <div className="flex flex-wrap items-center gap-1 text-xs text-muted-foreground">
                      <span>{t("exchangeRate.source", { source: t("exchangeRate.nobitex") })}</span>
                      <span aria-hidden="true">•</span>
                      <span>{t("exchangeRate.syncedAt", { time: syncedAt })}</span>
                    </div>
                  </div>

                  {approximateUSD ? (
                    <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3">
                      <span className="text-sm text-muted-foreground">{t("exchangeRate.estimatedUSD")}</span>
                      <strong className="text-base font-semibold tabular-nums text-foreground" dir="ltr">
                        {approximateUSD}
                      </strong>
                    </div>
                  ) : null}

                  {exchangeRate.status === "stale" ? (
                    <p className="text-xs leading-5 text-warning">{t("exchangeRate.stale")}</p>
                  ) : null}
                </>
              ) : (
                <p className="text-xs leading-5 text-muted-foreground">{t("exchangeRate.unavailable")}</p>
              )}
            </div>
          </Field>

          <div className="flex flex-col gap-3">
            <span className="inline-flex items-center gap-2 text-sm font-medium text-foreground">
              <Landmark className="size-4 text-primary" />
              {t("providerLabel")}
            </span>
            <RadioGroup
              value={provider}
              onValueChange={setProvider}
              className="grid grid-cols-3 gap-3"
            >
              {providerOptions.map((option) => {
                const Icon = option.icon;
                return (
                  <Label
                    key={option.value}
                    htmlFor={`topup-provider-${option.value}`}
                    className={cn(
                      "flex min-w-0 cursor-pointer flex-col items-center gap-3 rounded-xl border border-border bg-card px-3 py-4 text-center transition-colors",
                      provider === option.value && "border-primary bg-accent/40 text-accent-foreground",
                    )}
                  >
                    <div className="flex min-w-0 flex-1 flex-col items-center gap-3">
                      <div className="flex size-14 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                        {option.logoSrc ? (
                          <Image src={option.logoSrc} alt={option.label} width={56} height={56} className="size-14 object-contain" />
                        ) : (
                          <Icon className="size-14" />
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="text-sm font-medium text-foreground">{option.label}</div>
                        <p className="mt-1 text-xs leading-5 text-muted-foreground">{option.description}</p>
                      </div>
                    </div>
                    <RadioGroupItem id={`topup-provider-${option.value}`} value={option.value} />
                  </Label>
                );
              })}
            </RadioGroup>
          </div>

          <div className="inline-flex items-center gap-2 text-xs leading-5 text-muted-foreground">
            <BadgePercent className="size-4 text-primary" />
            <span>{t("taxNotice")}</span>
          </div>

          <div className="flex justify-end">
            <Button type="submit" disabled={loading}>
              <CreditCard data-icon="inline-start" />
              {loading ? t("submitLoading") : t("submit")}
            </Button>
          </div>
        </form>
      </SettingsCard>
    </div>
  );
}

function parseCurrentExchangeRate(value: unknown): CurrentExchangeRate | null {
  if (!value || typeof value !== "object") return null;

  const candidate = value as Partial<CurrentExchangeRate>;
  const fetchedAt = typeof candidate.fetchedAt === "string" ? new Date(candidate.fetchedAt) : null;
  const source = typeof candidate.source === "string" ? candidate.source.trim().toLowerCase() : "";
  if (
    candidate.base !== "USD" ||
    candidate.quote !== "IRR" ||
    typeof candidate.rate !== "number" ||
    !Number.isFinite(candidate.rate) ||
    candidate.rate <= 0 ||
    !source.startsWith("nobitex:") ||
    !fetchedAt ||
    Number.isNaN(fetchedAt.getTime()) ||
    (candidate.status !== "fresh" && candidate.status !== "stale")
  ) {
    return null;
  }

  return candidate as CurrentExchangeRate;
}
