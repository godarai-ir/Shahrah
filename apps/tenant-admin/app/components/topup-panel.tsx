"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { BadgePercent, CreditCard, Landmark, TestTubeDiagonal, WalletCards } from "lucide-react";

import { SettingsCard } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { cn } from "@/lib/utils";

type TopUpIntentResponse = {
  payment_url?: string;
};

type Props = {
  initialError: string | null;
};

export default function TopUpPanel(props: Props) {
  const t = useTranslations("topupPage");

  const [amountToman, setAmountToman] = useState("");
  const [provider, setProvider] = useState("mock");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(props.initialError);

  const quickAmounts = [
    { value: "1000000", label: t("quickAmounts.oneMillion") },
    { value: "3000000", label: t("quickAmounts.threeMillion") },
    { value: "10000000", label: t("quickAmounts.tenMillion") },
    { value: "50000000", label: t("quickAmounts.fiftyMillion") },
  ];

  const providerOptions = [
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
  ] as const;

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
          <div className="flex flex-col gap-2">
            <Label htmlFor="amountToman">{t("amountLabel")}</Label>
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
          </div>

          <div className="flex flex-col gap-3">
            <span className="inline-flex items-center gap-2 text-sm font-medium text-foreground">
              <Landmark className="size-4 text-primary" />
              {t("providerLabel")}
            </span>
            <RadioGroup
              dir="rtl"
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
