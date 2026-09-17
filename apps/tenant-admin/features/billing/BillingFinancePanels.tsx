"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { BarChart3, Building2, CreditCard, Download, FileImage, FileText, Landmark, ReceiptText, Tag, Trash2, Upload, WalletCards, type LucideIcon } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { FormSection } from "@/components/app-ui/FormSection";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Attachment,
  AttachmentAction,
  AttachmentActions,
  AttachmentContent,
  AttachmentDescription,
  AttachmentMedia,
  AttachmentTitle,
} from "@/components/ui/attachment";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  type BankTransferRequest,
  type BillingProfile,
  type BillingProviderUsage,
  type BillingSummaryPreset,
  type BillingUsageSummary,
  type DiscountRedemption,
  type Invoice,
  type MonthlyUsageDetail,
  type WalletTransaction,
  buildRecentJalaliMonthOptions,
  getCurrentJalaliDateValue,
  getCurrentJalaliMonthSelection,
  getJalaliMonthKey,
  normalizeJalaliDateInput,
  parseJalaliMonthKey,
  statusTone,
} from "@/features/billing/domain";
import { getBillingReports, submitBankTransfer, updateBillingProfile } from "@/features/billing/api";
import {
  buildFilterOptions,
  formatFileSize,
  JalaliDateField,
  mapDiscountTypeLabel,
  mapDocumentStatusLabel,
  mapTransactionTypeLabel,
  matchDiscountFilters,
  matchInvoiceFilters,
  matchTransactionFilters,
  normalizeDigits,
  ProfileTypeField,
  transactionTypeIcon,
  type BillingProfileType,
} from "@/features/billing/billing-ui-helpers";
import { downloadBillingDocumentsCSV, type BillingDocumentTab } from "@/features/billing/documents-csv";
import { useBillingFormat } from "@/features/billing/use-billing-format";
import { useAppLocale } from "@/hooks/use-locale-format";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
};

const REPORT_PRESETS: BillingSummaryPreset[] = [
  "current_month",
  "current_quarter",
  "current_year",
  "last_3_months",
  "last_6_months",
  "last_12_months",
];
const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_RECEIPT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const BANK_TRANSFER_ACCOUNT_INFO = {
  iban: "IR50 0560 9701 8000 0782 3380 01",
  cardNumber: "6219 8610 8689 9761",
} as const;

const ProviderCostChart = dynamic(() => import("@/features/billing/provider-cost-chart"), {
  loading: () => <div className="h-full w-full animate-pulse rounded-md bg-muted" aria-hidden="true" />,
});

export function BankTransferPanel({ initialItems, initialError }: BankTransferPanelProps) {
  const router = useRouter();
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const [amountToman, setAmountToman] = useState("");
  const [trackingNumber, setTrackingNumber] = useState("");
  const [bankTransactionDateJalali, setBankTransactionDateJalali] = useState(getCurrentJalaliDateValue());
  const [receiptImage, setReceiptImage] = useState<File | null>(null);
  const [receiptPreviewURL, setReceiptPreviewURL] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(initialError);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    return () => {
      if (receiptPreviewURL) {
        URL.revokeObjectURL(receiptPreviewURL);
      }
    };
  }, [receiptPreviewURL]);

  const columns = useMemo<ColumnDef<BankTransferRequest>[]>(
    () => [
      { accessorKey: "tracking_number", header: t("bankTransfer.columns.trackingNumber") },
      { accessorKey: "amount_irr", header: t("bankTransfer.columns.amount"), cell: ({ row }) => format.toman(row.original.amount_irr) },
      { accessorKey: "bank_transaction_at", header: t("bankTransfer.columns.transferDate"), cell: ({ row }) => format.date(row.original.bank_transaction_at) },
      { accessorKey: "status", header: t("bankTransfer.columns.status"), cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status, t)}</StatusBadge> },
      { accessorKey: "created_at", header: t("bankTransfer.columns.createdAt"), cell: ({ row }) => format.dateTime(row.original.created_at) },
    ],
    [format, t],
  );

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const toman = Number(amountToman.replaceAll(",", ""));
    if (!Number.isFinite(toman) || toman <= 0 || !trackingNumber.trim()) {
      setError(t("bankTransfer.errors.amountAndTracking"));
      return;
    }
    const normalizedDate = normalizeJalaliDateInput(bankTransactionDateJalali);
    if (!/^\d{4}-\d{2}-\d{2}$/.test(normalizedDate)) {
      setError(t("bankTransfer.errors.jalaliDate"));
      return;
    }
    if (!receiptImage) {
      setError(t("bankTransfer.errors.receiptRequired"));
      return;
    }

    setLoading(true);
    try {
      const formData = new FormData();
      formData.set("amount_irr", String(Math.round(toman * 10)));
      formData.set("tracking_number", trackingNumber.trim());
      formData.set("bank_transaction_date_jalali", normalizedDate);
      formData.set("receipt_image", receiptImage);

      await submitBankTransfer(formData);
      router.refresh();
      setAmountToman("");
      setTrackingNumber("");
      setBankTransactionDateJalali(getCurrentJalaliDateValue());
      if (receiptPreviewURL) {
        URL.revokeObjectURL(receiptPreviewURL);
      }
      setReceiptPreviewURL(null);
      setReceiptImage(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("unknownError"));
    } finally {
      setLoading(false);
    }
  }

  function handleReceiptChange(e: ChangeEvent<HTMLInputElement>) {
    const nextFile = e.target.files?.[0] ?? null;
    e.target.value = "";

    if (!nextFile) return;
    if (!ALLOWED_RECEIPT_IMAGE_TYPES.includes(nextFile.type)) {
      setError(t("bankTransfer.errors.receiptType"));
      return;
    }
    if (nextFile.size > MAX_RECEIPT_IMAGE_SIZE) {
      setError(t("bankTransfer.errors.receiptSize"));
      return;
    }

    if (receiptPreviewURL) {
      URL.revokeObjectURL(receiptPreviewURL);
    }
    setError(null);
    setReceiptImage(nextFile);
    setReceiptPreviewURL(URL.createObjectURL(nextFile));
  }

  function clearReceipt() {
    if (receiptPreviewURL) {
      URL.revokeObjectURL(receiptPreviewURL);
    }
    setReceiptPreviewURL(null);
    setReceiptImage(null);
  }

  return (
    <div className="flex flex-col gap-6">
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <BankAccountDetailsCard />
      <SettingsCard>
        <form onSubmit={submit} className="grid gap-4 md:grid-cols-2">
          <AppTextField id="bank-amount" label={t("bankTransfer.form.amount")} value={amountToman} onChange={(event) => setAmountToman(event.target.value)} dir="ltr" inputMode="numeric" />
          <AppTextField id="bank-tracking" label={t("bankTransfer.form.trackingNumber")} value={trackingNumber} onChange={(event) => setTrackingNumber(event.target.value)} dir="ltr" />
          <JalaliDateField id="bank-date" label={t("bankTransfer.form.transferDate")} value={bankTransactionDateJalali} onChange={setBankTransactionDateJalali} />
          <div className="flex flex-col gap-2 md:col-span-2">
            <Label htmlFor="bank-receipt">{t("bankTransfer.form.receipt")}</Label>
            <div className="flex flex-col gap-3">
              <input
                id="bank-receipt"
                type="file"
                accept="image/jpeg,image/png,image/webp,image/gif"
                className="sr-only"
                onChange={handleReceiptChange}
              />
              {receiptImage ? (
                <Attachment orientation="vertical" state={loading ? "uploading" : "done"}>
                  <AttachmentMedia variant="image">
                    {receiptPreviewURL ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={receiptPreviewURL} alt={receiptImage.name} />
                    ) : (
                      <FileImage aria-hidden="true" />
                    )}
                  </AttachmentMedia>
                  <AttachmentContent>
                    <AttachmentTitle>{receiptImage.name}</AttachmentTitle>
                    <AttachmentDescription>
                      {formatFileSize(receiptImage.size)} · {t("bankTransfer.form.receiptReady")}
                    </AttachmentDescription>
                  </AttachmentContent>
                  <AttachmentActions>
                    <AttachmentAction aria-label={t("bankTransfer.form.removeReceipt", { name: receiptImage.name })} onClick={clearReceipt} disabled={loading}>
                      <Trash2 />
                    </AttachmentAction>
                  </AttachmentActions>
                </Attachment>
              ) : (
                <label
                  htmlFor="bank-receipt"
                  className="flex cursor-pointer items-center gap-3 rounded-xl border border-dashed border-border bg-muted/30 px-4 py-4 text-sm text-muted-foreground transition-colors hover:border-primary/50 hover:bg-accent"
                >
                  <Upload data-icon="inline-start" />
                  <span>{t("bankTransfer.form.pickReceipt")}</span>
                </label>
              )}
              <p className="text-xs leading-5 text-muted-foreground">
                {t("bankTransfer.form.receiptHint")}
              </p>
            </div>
          </div>
          <div className="md:col-span-2">
            <Button type="submit" disabled={loading}>{loading ? t("bankTransfer.form.submitting") : t("bankTransfer.form.submit")}</Button>
          </div>
        </form>
      </SettingsCard>
      <SettingsCard title={t("bankTransfer.list.title")} description={t("bankTransfer.list.description")}>
        <DataTable tableId="billing-bank-transfers" columns={columns} data={initialItems} emptyLabel={t("bankTransfer.list.empty")} />
      </SettingsCard>
    </div>
  );
}

function BankAccountDetailsCard() {
  const t = useTranslations("billing");

  return (
    <section className="overflow-hidden rounded-xl border border-primary/30 bg-primary/5 text-card-foreground shadow-sm">
      <div className="flex flex-col gap-4 p-4 md:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 items-start gap-3">
            <span className="flex size-11 shrink-0 items-center justify-center rounded-lg border border-primary/20 bg-background text-primary">
              <Landmark className="size-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <h2 className="text-base font-semibold md:text-lg">{t("bankTransfer.account.title")}</h2>
              <p className="text-sm leading-6 text-muted-foreground">
                {t("bankTransfer.account.description")}
              </p>
            </div>
          </div>
          <div className="inline-flex w-fit items-center gap-2 rounded-lg border border-primary/20 bg-background px-3 py-2 text-sm font-semibold text-primary">
            <Building2 className="size-4" aria-hidden="true" />
            {t("bankTransfer.account.bankName")}
          </div>
        </div>
        <div className="grid gap-3 lg:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)]">
          <BankAccountCopyField
            icon={Landmark}
            label={t("bankTransfer.account.ibanLabel")}
            value={BANK_TRANSFER_ACCOUNT_INFO.iban}
            valueAriaLabel={t("bankTransfer.account.ibanLabel")}
          />
          <BankAccountCopyField
            icon={CreditCard}
            label={t("bankTransfer.account.cardLabel")}
            value={BANK_TRANSFER_ACCOUNT_INFO.cardNumber}
            valueAriaLabel={t("bankTransfer.account.cardValueAriaLabel")}
          />
        </div>
      </div>
    </section>
  );
}

function BankAccountCopyField({
  icon: Icon,
  label,
  value,
  valueAriaLabel,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  valueAriaLabel: string;
}) {
  const t = useTranslations("billing");

  return (
    <div className="grid gap-2 rounded-lg border border-border bg-background p-3">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Icon className="size-4 text-primary" aria-hidden="true" />
        <span>{label}</span>
      </div>
      <CopyableValue
        value={value}
        labels={{
          valueAriaLabel,
          copy: t("bankTransfer.account.copy"),
          copying: t("bankTransfer.account.copying"),
          copied: t("bankTransfer.account.copied"),
        }}
      />
    </div>
  );
}

type ReportsPanelProps = {
  initialSummary: BillingUsageSummary | null;
  initialMonthly: MonthlyUsageDetail | null;
  initialError: string | null;
};

export function ReportsPanel({ initialSummary, initialMonthly, initialError }: ReportsPanelProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const locale = useAppLocale();
  const [preset, setPreset] = useState<BillingSummaryPreset>(initialSummary?.preset ?? "current_month");
  const [monthKey, setMonthKey] = useState(() => {
    if (initialMonthly) {
      return getJalaliMonthKey({ year: initialMonthly.jalali_year, month: initialMonthly.jalali_month });
    }
    return getJalaliMonthKey(getCurrentJalaliMonthSelection());
  });
  const [summary, setSummary] = useState<BillingUsageSummary | null>(initialSummary);
  const [monthly, setMonthly] = useState<MonthlyUsageDetail | null>(initialMonthly);
  const [error, setError] = useState<string | null>(initialError);
  const [loading, setLoading] = useState(false);
  const monthOptions = useMemo(() => buildRecentJalaliMonthOptions(18, new Date(), locale), [locale]);

  const columns = useMemo<ColumnDef<BillingProviderUsage>[]>(
    () => [
      { accessorKey: "provider_name", header: t("reports.columns.provider") },
      { accessorKey: "request_count", header: t("reports.columns.requests") },
      { accessorKey: "input_tokens", header: t("reports.columns.inputTokens") },
      { accessorKey: "output_tokens", header: t("reports.columns.outputTokens") },
      { accessorKey: "raw_cost_irr", header: t("reports.columns.rawCost"), cell: ({ row }) => format.irr(row.original.raw_cost_irr) },
      { accessorKey: "discount_irr", header: t("reports.columns.discount"), cell: ({ row }) => format.irr(row.original.discount_irr) },
      { accessorKey: "cache_savings_irr", header: t("reports.columns.cacheSavings"), cell: ({ row }) => format.irr(row.original.cache_savings_irr) },
      { accessorKey: "final_cost_irr", header: t("reports.columns.finalCost"), cell: ({ row }) => format.irr(row.original.final_cost_irr) },
    ],
    [format, t],
  );

  async function refresh(nextPreset = preset, nextMonthKey = monthKey) {
    const nextMonth = parseJalaliMonthKey(nextMonthKey);
    if (!nextMonth) {
      setError(t("reports.invalidMonth"));
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const reports = await getBillingReports(nextPreset, nextMonth);
      setSummary(reports.summary);
      setMonthly(reports.monthly);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("unknownError"));
    } finally {
      setLoading(false);
    }
  }

  const pieData = (summary?.providers ?? []).map((item) => ({
    name: item.provider_name || "unknown",
    value: Math.max(item.final_cost_irr, 0),
  }));

  return (
    <div className="space-y-6">
      {error ? <Alert variant="destructive"><AlertTitle>{t("errorTitle")}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        <MetricCard icon={WalletCards} label={t("reports.metrics.periodUsage")} value={format.irr(summary?.total_usage_irr)} />
        <MetricCard icon={Tag} label={t("reports.metrics.discount")} value={format.irr(summary?.discount_irr)} />
        <MetricCard icon={BarChart3} label={t("reports.metrics.cacheSavings")} value={format.irr(summary?.cache_savings_irr)} />
        <MetricCard icon={ReceiptText} label={t("reports.metrics.giftCredit")} value={format.irr(summary?.gift_credit_irr)} />
      </div>
      <SettingsCard title={t("reports.filter.title")} description={t("reports.filter.description")}>
        <div className="grid gap-4 md:grid-cols-3">
          <AppSelectField
            label={t("reports.filter.range")}
            value={preset}
            onValueChange={(value) => {
              setPreset(value as BillingSummaryPreset);
              void refresh(value as BillingSummaryPreset, monthKey);
            }}
            options={REPORT_PRESETS.map((value) => ({ value, label: t(`reports.presets.${value}`) }))}
          />
          <AppSelectField
            label={t("reports.filter.monthly")}
            value={monthKey}
            onValueChange={setMonthKey}
            options={monthOptions}
          />
          <div className="flex items-end">
            <Button type="button" onClick={() => void refresh()} disabled={loading}>{loading ? t("reports.filter.refreshing") : t("reports.filter.refresh")}</Button>
          </div>
        </div>
      </SettingsCard>
      <div className="grid min-w-0 gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <SettingsCard
          title={t("reports.monthlyDetail.title")}
          description={t("reports.monthlyDetail.description", { month: format.jalaliMonth(monthKey) })}
          className="min-w-0 max-w-none overflow-hidden"
        >
          <DataTable tableId="billing-monthly-providers" columns={columns} data={monthly?.providers ?? []} emptyLabel={t("reports.monthlyDetail.empty")} />
        </SettingsCard>
        <AppCard title={t("reports.providerShare.title")} description={t("reports.providerShare.description")}>
          <div className="h-64">
            <ProviderCostChart data={pieData} />
          </div>
        </AppCard>
      </div>
    </div>
  );
}

type ProfilePanelProps = {
  initialProfile: BillingProfile | null;
  initialError: string | null;
};

export function BillingProfilePanel({ initialProfile, initialError }: ProfilePanelProps) {
  const router = useRouter();
  const t = useTranslations("billing");
  const data = initialProfile?.data ?? {};
  const initialProfileType: BillingProfileType = initialProfile?.profile_type === "legal" ? "legal" : "individual";
  const [profileType, setProfileType] = useState<BillingProfileType>(initialProfileType);
  const [displayName, setDisplayName] = useState(String(data.display_name ?? ""));
  const [nationalID, setNationalID] = useState(String(data.national_id ?? ""));
  const [economicCode, setEconomicCode] = useState(String(data.economic_code ?? ""));
  const [postalCode, setPostalCode] = useState(String(data.postal_code ?? ""));
  const [address, setAddress] = useState(String(data.address ?? ""));
  const [phone, setPhone] = useState(String(data.phone ?? ""));
  const [email, setEmail] = useState(String(data.email ?? ""));
  const [error, setError] = useState<string | null>(initialError);
  const [loading, setLoading] = useState(false);

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    const normalizedNationalID = normalizeDigits(nationalID);
    const normalizedEconomicCode = normalizeDigits(economicCode);
    const normalizedPostalCode = normalizeDigits(postalCode);
    const normalizedEmail = email.trim();
    if (!displayName.trim() || !normalizedNationalID) {
      setError(t("profile.errors.displayNameAndID"));
      return;
    }
    if (profileType === "individual" && normalizedNationalID.length !== 10) {
      setError(t("profile.errors.nationalID"));
      return;
    }
    if (profileType === "legal" && normalizedNationalID.length !== 11) {
      setError(t("profile.errors.legalID"));
      return;
    }
    if (profileType === "legal" && !normalizedEconomicCode) {
      setError(t("profile.errors.economicCodeRequired"));
      return;
    }
    if (profileType === "legal" && normalizedEconomicCode.length !== 12) {
      setError(t("profile.errors.economicCode"));
      return;
    }
    if (postalCode.trim() && normalizedPostalCode.length !== 10) {
      setError(t("profile.errors.postalCode"));
      return;
    }
    if (normalizedEmail && !EMAIL_PATTERN.test(normalizedEmail)) {
      setError(t("profile.errors.email"));
      return;
    }
    setLoading(true);
    try {
      await updateBillingProfile({
        profile_type: profileType,
        data: {
          display_name: displayName.trim(),
          national_id: normalizedNationalID,
          economic_code: normalizedEconomicCode,
          postal_code: normalizedPostalCode,
          address: address.trim(),
          phone: phone.trim(),
          email: normalizedEmail,
        },
      });
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("unknownError"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {error ? <Alert variant="destructive"><AlertTitle>{t("errorTitle")}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <form onSubmit={submit} className="flex flex-col gap-4">
        <FormSection title={t("profile.identity.title")} description={t("profile.identity.description")}>
          <ProfileTypeField value={profileType} onValueChange={setProfileType} />
          <DisplayNameField id="profile-name" label={t("profile.identity.displayName")} value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
          <AppTextField
            id="profile-national"
            label={profileType === "legal" ? t("profile.identity.legalID") : t("profile.identity.nationalID")}
            value={nationalID}
            dir="ltr"
            inputMode="numeric"
            maxLength={profileType === "legal" ? 11 : 10}
            onChange={(event) => setNationalID(normalizeDigits(event.target.value).slice(0, profileType === "legal" ? 11 : 10))}
          />
          {profileType === "legal" ? (
            <AppTextField
              id="profile-economic"
              label={t("profile.identity.economicCode")}
              value={economicCode}
              dir="ltr"
              inputMode="numeric"
              maxLength={12}
              onChange={(event) => setEconomicCode(normalizeDigits(event.target.value).slice(0, 12))}
            />
          ) : null}
        </FormSection>
        <FormSection title={t("profile.contact.title")} description={t("profile.contact.description")}>
          <AppTextField
            id="profile-postal"
            label={t("profile.contact.postalCode")}
            value={postalCode}
            dir="ltr"
            inputMode="numeric"
            maxLength={10}
            onChange={(event) => setPostalCode(normalizeDigits(event.target.value).slice(0, 10))}
          />
          <AppTextField id="profile-address" label={t("profile.contact.address")} value={address} onChange={(event) => setAddress(event.target.value)} />
          <AppTextField id="profile-phone" label={t("profile.contact.phone")} value={phone} onChange={(event) => setPhone(event.target.value)} />
          <AppTextField
            id="profile-email"
            label={t("profile.contact.email")}
            value={email}
            type="email"
            dir="ltr"
            inputMode="email"
            autoComplete="email"
            onChange={(event) => setEmail(event.target.value)}
          />
        </FormSection>
        <Button type="submit" disabled={loading} className="self-end">{loading ? t("profile.saving") : t("profile.save")}</Button>
      </form>
    </div>
  );
}

type DocumentsPanelProps = {
  invoices: Invoice[];
  transactions: WalletTransaction[];
  discounts: DiscountRedemption[];
};

export function BillingDocumentsPanel({ invoices, transactions, discounts }: DocumentsPanelProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const allLabel = t("filters.all");
  const [activeTab, setActiveTab] = useState<BillingDocumentTab>("invoices");
  const [invoiceStatus, setInvoiceStatus] = useState("all");
  const [invoicePage, setInvoicePage] = useState(1);
  const [transactionStatus, setTransactionStatus] = useState("all");
  const [transactionType, setTransactionType] = useState("all");
  const [transactionPage, setTransactionPage] = useState(1);
  const [discountStatus, setDiscountStatus] = useState("all");
  const [discountType, setDiscountType] = useState("all");
  const [discountPage, setDiscountPage] = useState(1);

  const invoiceColumns = useMemo<ColumnDef<Invoice>[]>(
    () => [
      { accessorKey: "net_adjustment_irr", header: t("documents.invoices.netAdjustment"), cell: ({ row }) => format.amount(row.original.net_adjustment_irr) },
      { accessorKey: "final_amount_irr", header: t("documents.invoices.finalAmount"), cell: ({ row }) => format.amount(row.original.final_amount_irr) },
      { accessorKey: "period_end", header: t("documents.invoices.periodEnd"), cell: ({ row }) => format.dateTime(row.original.period_end) },
      { accessorKey: "status", header: t("documents.invoices.status"), cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status, t)}</StatusBadge> },
      { accessorKey: "invoice_number", header: t("documents.invoices.number") },
    ],
    [format, t],
  );
  const txColumns = useMemo<ColumnDef<WalletTransaction>[]>(
    () => [
      {
        accessorKey: "type",
        header: t("documents.transactions.type"),
        cell: ({ row }) => (
          <div className="flex items-center gap-2">
            {transactionTypeIcon(row.original.type)}
            <span>{mapTransactionTypeLabel(row.original.type, t)}</span>
          </div>
        ),
      },
      { accessorKey: "status", header: t("documents.transactions.status"), cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status, t)}</StatusBadge> },
      { accessorKey: "amount_irr", header: t("documents.transactions.amount"), cell: ({ row }) => format.amount(row.original.amount_irr) },
      { accessorKey: "balance_after_irr", header: t("documents.transactions.balanceAfter"), cell: ({ row }) => format.amount(row.original.balance_after_irr) },
      { accessorKey: "created_at", header: t("documents.transactions.createdAt"), cell: ({ row }) => format.dateTime(row.original.created_at) },
    ],
    [format, t],
  );
  const discountColumns = useMemo<ColumnDef<DiscountRedemption>[]>(
    () => [
      { accessorKey: "code", header: t("documents.discounts.code") },
      { accessorKey: "discount_type", header: t("documents.discounts.type"), cell: ({ row }) => mapDiscountTypeLabel(row.original.discount_type, t) },
      { accessorKey: "status", header: t("documents.discounts.status"), cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status, t)}</StatusBadge> },
      { accessorKey: "used_amount_irr", header: t("documents.discounts.usedAmount"), cell: ({ row }) => format.amount(row.original.used_amount_irr) },
      { accessorKey: "created_at", header: t("documents.discounts.createdAt"), cell: ({ row }) => format.dateTime(row.original.created_at) },
    ],
    [format, t],
  );

  const invoiceRows = useMemo(
    () => invoices.filter((item) => matchInvoiceFilters(item, invoiceStatus)),
    [invoiceStatus, invoices],
  );
  const transactionRows = useMemo(
    () => transactions.filter((item) => matchTransactionFilters(item, transactionStatus, transactionType)),
    [transactionStatus, transactionType, transactions],
  );
  const discountRows = useMemo(
    () => discounts.filter((item) => matchDiscountFilters(item, discountStatus, discountType)),
    [discountStatus, discountType, discounts],
  );

  return (
    <AppCard className="min-w-0 max-w-none overflow-hidden">
      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as BillingDocumentTab)} className="min-w-0 w-full">
        <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="shrink-0 self-start"
            aria-label={t("documents.exportCSV")}
            title={t("documents.exportCSV")}
            onClick={() =>
              downloadBillingDocumentsCSV({
                tab: activeTab,
                invoices: invoiceRows,
                transactions: transactionRows,
                discounts: discountRows,
                t,
                format,
              })
            }
          >
            <Download className="size-4" />
          </Button>
          <TabsList className="w-full justify-start overflow-x-auto sm:w-fit sm:self-end">
            <TabsTrigger value="invoices"><FileText className="me-2 size-4" />{t("documents.tabs.invoices")}</TabsTrigger>
            <TabsTrigger value="transactions"><ReceiptText className="me-2 size-4" />{t("documents.tabs.transactions")}</TabsTrigger>
            <TabsTrigger value="discounts"><Tag className="me-2 size-4" />{t("documents.tabs.discounts")}</TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="invoices">
          <DocumentTableSection
            columns={invoiceColumns}
            rows={invoiceRows}
            emptyLabel={t("documents.invoices.empty")}
            filters={[
              {
                value: invoiceStatus,
                onValueChange: (value) => {
                  setInvoiceStatus(value);
                  setInvoicePage(1);
                },
                options: buildFilterOptions(invoices.map((item) => item.status), (value) => mapDocumentStatusLabel(value, t), allLabel),
                label: t("documents.invoices.status"),
              },
            ]}
            page={invoicePage}
            onPageChange={setInvoicePage}
          />
        </TabsContent>
        <TabsContent value="transactions">
          <DocumentTableSection
            columns={txColumns}
            rows={transactionRows}
            emptyLabel={t("documents.transactions.empty")}
            filters={[
              {
                value: transactionType,
                onValueChange: (value) => {
                  setTransactionType(value);
                  setTransactionPage(1);
                },
                options: buildFilterOptions(transactions.map((item) => item.type), (value) => mapTransactionTypeLabel(value, t), allLabel),
                label: t("documents.transactions.type"),
              },
              {
                value: transactionStatus,
                onValueChange: (value) => {
                  setTransactionStatus(value);
                  setTransactionPage(1);
                },
                options: buildFilterOptions(transactions.map((item) => item.status), (value) => mapDocumentStatusLabel(value, t), allLabel),
                label: t("documents.transactions.status"),
              },
            ]}
            page={transactionPage}
            onPageChange={setTransactionPage}
          />
        </TabsContent>
        <TabsContent value="discounts">
          <DocumentTableSection
            columns={discountColumns}
            rows={discountRows}
            emptyLabel={t("documents.discounts.empty")}
            filters={[
              {
                value: discountType,
                onValueChange: (value) => {
                  setDiscountType(value);
                  setDiscountPage(1);
                },
                options: buildFilterOptions(discounts.map((item) => item.discount_type), (value) => mapDiscountTypeLabel(value, t), allLabel),
                label: t("documents.discounts.type"),
              },
              {
                value: discountStatus,
                onValueChange: (value) => {
                  setDiscountStatus(value);
                  setDiscountPage(1);
                },
                options: buildFilterOptions(discounts.map((item) => item.status), (value) => mapDocumentStatusLabel(value, t), allLabel),
                label: t("documents.discounts.status"),
              },
            ]}
            page={discountPage}
            onPageChange={setDiscountPage}
          />
        </TabsContent>
      </Tabs>
    </AppCard>
  );
}

type DocumentFilterControl = {
  label: string;
  value: string;
  onValueChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
};

function DocumentTableSection<T>({
  columns,
  rows,
  emptyLabel,
  filters,
  page,
  onPageChange,
}: {
  columns: ColumnDef<T>[];
  rows: T[];
  emptyLabel: string;
  filters: DocumentFilterControl[];
  page: number;
  onPageChange: (value: number) => void;
}) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const pageSizeNumber = 10;
  const pageCount = Math.max(1, Math.ceil(rows.length / pageSizeNumber));
  const safePage = Math.min(page, pageCount);
  const pageStart = rows.length === 0 ? 0 : (safePage - 1) * pageSizeNumber;
  const paginatedRows = rows.slice(pageStart, pageStart + pageSizeNumber);
  const pageEnd = rows.length === 0 ? 0 : Math.min(pageStart + pageSizeNumber, rows.length);

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex flex-col gap-3 rounded-xl border border-border bg-muted/20 p-3">
        <div className="flex flex-col gap-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            {filters.map((filter) => (
              <AppSelectField
                key={filter.label}
                value={filter.value}
                onValueChange={filter.onValueChange}
                options={filter.options}
                label={filter.label}
                triggerClassName="min-w-36"
              />
            ))}
          </div>
        </div>
      </div>
      <DataTable tableId="billing-report-rows" columns={columns} data={paginatedRows} emptyLabel={emptyLabel} />
      <div className="flex flex-col gap-3 border-t border-border pt-4 text-sm sm:flex-row sm:items-center sm:justify-between">
        <div className="text-muted-foreground">
          {t("pagination.range", {
            from: format.count(pageStart + 1),
            to: format.count(pageEnd),
            total: format.count(rows.length),
          })}
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" disabled={safePage <= 1} onClick={() => onPageChange(safePage - 1)}>
            {t("pagination.previous")}
          </Button>
          <span className="min-w-20 text-center text-xs text-muted-foreground">
            {t("pagination.page", { page: format.count(safePage), pageCount: format.count(pageCount) })}
          </span>
          <Button type="button" variant="outline" size="sm" disabled={safePage >= pageCount} onClick={() => onPageChange(safePage + 1)}>
            {t("pagination.next")}
          </Button>
        </div>
      </div>
    </div>
  );
}

