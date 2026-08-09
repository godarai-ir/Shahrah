"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
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
  formatDate,
  formatDateOnly,
  formatIRR,
  formatSelectedJalaliMonth,
  formatToman,
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
  downloadCSV,
  formatAmountValue,
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
  toPersianNumber,
  transactionTypeIcon,
  type BillingProfileType,
} from "@/features/billing/billing-ui-helpers";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
};

const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_RECEIPT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const BANK_TRANSFER_ACCOUNT_INFO = {
  bankName: "بانک سامان",
  iban: "IR50 0560 9701 8000 0782 3380 01",
  cardNumber: "6219 8610 8689 9761",
} as const;

const ProviderCostChart = dynamic(() => import("@/features/billing/provider-cost-chart"), {
  loading: () => <div className="h-full w-full animate-pulse rounded-md bg-muted" aria-hidden="true" />,
});

export function BankTransferPanel({ initialItems, initialError }: BankTransferPanelProps) {
  const router = useRouter();
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
      { accessorKey: "tracking_number", header: "شماره پیگیری" },
      { accessorKey: "amount_irr", header: "مبلغ", cell: ({ row }) => formatToman(row.original.amount_irr) },
      { accessorKey: "bank_transaction_at", header: "تاریخ واریز", cell: ({ row }) => formatDateOnly(row.original.bank_transaction_at) },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status)}</StatusBadge> },
      { accessorKey: "created_at", header: "ثبت", cell: ({ row }) => formatDate(row.original.created_at) },
    ],
    [],
  );

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const toman = Number(amountToman.replaceAll(",", ""));
    if (!Number.isFinite(toman) || toman <= 0 || !trackingNumber.trim()) {
      setError("مبلغ و شماره پیگیری را کامل وارد کنید.");
      return;
    }
    const normalizedDate = normalizeJalaliDateInput(bankTransactionDateJalali);
    if (!/^\d{4}-\d{2}-\d{2}$/.test(normalizedDate)) {
      setError("تاریخ واریز را با فرمت شمسی YYYY-MM-DD وارد کنید.");
      return;
    }
    if (!receiptImage) {
      setError("تصویر رسید را بارگذاری کنید.");
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
      setError(err instanceof Error ? err.message : "خطای ناشناخته رخ داد.");
    } finally {
      setLoading(false);
    }
  }

  function handleReceiptChange(e: ChangeEvent<HTMLInputElement>) {
    const nextFile = e.target.files?.[0] ?? null;
    e.target.value = "";

    if (!nextFile) return;
    if (!ALLOWED_RECEIPT_IMAGE_TYPES.includes(nextFile.type)) {
      setError("فقط فایل‌های تصویری رایج مانند JPG، PNG، WEBP و GIF مجاز هستند.");
      return;
    }
    if (nextFile.size > MAX_RECEIPT_IMAGE_SIZE) {
      setError("حجم تصویر رسید باید حداکثر ۲ مگابایت باشد.");
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
          <AlertTitle>خطا</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <BankAccountDetailsCard />
      <SettingsCard>
        <form onSubmit={submit} className="grid gap-4 md:grid-cols-2">
          <AppTextField id="bank-amount" label="مبلغ (تومان)" value={amountToman} onChange={(event) => setAmountToman(event.target.value)} dir="ltr" inputMode="numeric" />
          <AppTextField id="bank-tracking" label="شماره پیگیری" value={trackingNumber} onChange={(event) => setTrackingNumber(event.target.value)} dir="ltr" />
          <JalaliDateField id="bank-date" label="تاریخ واریز (شمسی)" value={bankTransactionDateJalali} onChange={setBankTransactionDateJalali} />
          <div className="flex flex-col gap-2 md:col-span-2">
            <Label htmlFor="bank-receipt">تصویر رسید</Label>
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
                      {formatFileSize(receiptImage.size)} · آماده برای ارسال
                    </AttachmentDescription>
                  </AttachmentContent>
                  <AttachmentActions>
                    <AttachmentAction aria-label={`حذف ${receiptImage.name}`} onClick={clearReceipt} disabled={loading}>
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
                  <span>یک تصویر رسید انتخاب کنید</span>
                </label>
              )}
              <p className="text-xs leading-5 text-muted-foreground">
                فقط یک تصویر با فرمت‌های JPG، PNG، WEBP یا GIF و حداکثر حجم ۲ مگابایت.
              </p>
            </div>
          </div>
          <div className="md:col-span-2">
            <Button type="submit" disabled={loading}>{loading ? "در حال ثبت..." : "ثبت و ارسال برای بررسی"}</Button>
          </div>
        </form>
      </SettingsCard>
      <SettingsCard title="واریزهای ثبت‌شده" description="آخرین درخواست‌های واریز بانکی">
        <DataTable columns={columns} data={initialItems} emptyLabel="درخواستی ثبت نشده است." />
      </SettingsCard>
    </div>
  );
}

function BankAccountDetailsCard() {
  return (
    <section className="overflow-hidden rounded-xl border border-primary/30 bg-primary/5 text-card-foreground shadow-sm">
      <div className="flex flex-col gap-4 p-4 md:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 items-start gap-3">
            <span className="flex size-11 shrink-0 items-center justify-center rounded-lg border border-primary/20 bg-background text-primary">
              <Landmark className="size-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <h2 className="text-base font-semibold md:text-lg">اطلاعات حساب برای واریز بانکی</h2>
              <p className="text-sm leading-6 text-muted-foreground">
                مبلغ را به حساب زیر واریز کنید و سپس رسید را برای بررسی بارگذاری کنید.
              </p>
            </div>
          </div>
          <div className="inline-flex w-fit items-center gap-2 rounded-lg border border-primary/20 bg-background px-3 py-2 text-sm font-semibold text-primary">
            <Building2 className="size-4" aria-hidden="true" />
            {BANK_TRANSFER_ACCOUNT_INFO.bankName}
          </div>
        </div>
        <div className="grid gap-3 lg:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)]">
          <BankAccountCopyField
            icon={Landmark}
            label="شماره شبا حساب بانکی"
            value={BANK_TRANSFER_ACCOUNT_INFO.iban}
            valueAriaLabel="شماره شبا حساب بانکی"
          />
          <BankAccountCopyField
            icon={CreditCard}
            label="شماره کارت"
            value={BANK_TRANSFER_ACCOUNT_INFO.cardNumber}
            valueAriaLabel="شماره کارت بانکی"
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
          copy: "کپی",
          copying: "در حال کپی",
          copied: "کپی شد",
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
  const monthOptions = useMemo(() => buildRecentJalaliMonthOptions(18), []);

  const columns = useMemo<ColumnDef<BillingProviderUsage>[]>(
    () => [
      { accessorKey: "provider_name", header: "ارائه‌دهنده" },
      { accessorKey: "request_count", header: "درخواست‌ها" },
      { accessorKey: "input_tokens", header: "توکن ورودی" },
      { accessorKey: "output_tokens", header: "توکن خروجی" },
      { accessorKey: "raw_cost_irr", header: "هزینه خام", cell: ({ row }) => formatIRR(row.original.raw_cost_irr) },
      { accessorKey: "discount_irr", header: "تخفیف", cell: ({ row }) => formatIRR(row.original.discount_irr) },
      { accessorKey: "cache_savings_irr", header: "صرفه‌جویی کش", cell: ({ row }) => formatIRR(row.original.cache_savings_irr) },
      { accessorKey: "final_cost_irr", header: "هزینه نهایی", cell: ({ row }) => formatIRR(row.original.final_cost_irr) },
    ],
    [],
  );

  async function refresh(nextPreset = preset, nextMonthKey = monthKey) {
    const nextMonth = parseJalaliMonthKey(nextMonthKey);
    if (!nextMonth) {
      setError("ماه گزارش باید یک ماه شمسی معتبر باشد.");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const reports = await getBillingReports(nextPreset, nextMonth);
      setSummary(reports.summary);
      setMonthly(reports.monthly);
    } catch (err) {
      setError(err instanceof Error ? err.message : "خطای ناشناخته رخ داد.");
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
      {error ? <Alert variant="destructive"><AlertTitle>خطا</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        <MetricCard icon={WalletCards} label="مصرف دوره" value={formatIRR(summary?.total_usage_irr)} />
        <MetricCard icon={Tag} label="تخفیف" value={formatIRR(summary?.discount_irr)} />
        <MetricCard icon={BarChart3} label="صرفه‌جویی کش" value={formatIRR(summary?.cache_savings_irr)} />
        <MetricCard icon={ReceiptText} label="اعتبار هدیه" value={formatIRR(summary?.gift_credit_irr)} />
      </div>
      <SettingsCard title="فیلتر گزارش" description="گزارش مصرف را در بازه زمانی دلخواه یا به تفکیک ماه مشاهده کنید.">
        <div className="grid gap-4 md:grid-cols-3">
          <AppSelectField
            label="بازه"
            value={preset}
            onValueChange={(value) => {
              setPreset(value as BillingSummaryPreset);
              void refresh(value as BillingSummaryPreset, monthKey);
            }}
            options={[
              { value: "current_month", label: "ماه جاری" },
              { value: "current_quarter", label: "فصل جاری" },
              { value: "current_year", label: "سال جاری" },
              { value: "last_3_months", label: "سه ماه اخیر" },
              { value: "last_6_months", label: "شش ماه اخیر" },
              { value: "last_12_months", label: "دوازده ماه اخیر" },
            ]}
          />
          <AppSelectField
            label="فیلتر ماهانه"
            value={monthKey}
            onValueChange={setMonthKey}
            options={monthOptions}
          />
          <div className="flex items-end">
            <Button type="button" onClick={() => void refresh()} disabled={loading}>{loading ? "در حال دریافت..." : "به‌روزرسانی"}</Button>
          </div>
        </div>
      </SettingsCard>
      <div className="grid min-w-0 gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <SettingsCard
          title="جزئیات ماهانه"
          description={`تجمیع مصرف، توکن و هزینه نهایی به تفکیک ارائه‌دهنده برای ${formatSelectedJalaliMonth(monthKey)}`}
          className="min-w-0 max-w-none overflow-hidden"
        >
          <DataTable columns={columns} data={monthly?.providers ?? []} emptyLabel="داده‌ای برای این ماه وجود ندارد." />
        </SettingsCard>
        <AppCard title="سهم ارائه‌دهنده‌ها" description="بر اساس هزینه نهایی">
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
      setError("نام نمایشی و شناسه ملی/کد ملی الزامی است.");
      return;
    }
    if (profileType === "individual" && normalizedNationalID.length !== 10) {
      setError("کد ملی باید دقیقاً ۱۰ رقم باشد.");
      return;
    }
    if (profileType === "legal" && normalizedNationalID.length !== 11) {
      setError("شناسه ملی باید دقیقاً ۱۱ رقم باشد.");
      return;
    }
    if (profileType === "legal" && !normalizedEconomicCode) {
      setError("برای پروفایل حقوقی، کد اقتصادی الزامی است.");
      return;
    }
    if (profileType === "legal" && normalizedEconomicCode.length !== 12) {
      setError("کد اقتصادی باید دقیقاً ۱۲ رقم باشد.");
      return;
    }
    if (postalCode.trim() && normalizedPostalCode.length !== 10) {
      setError("کد پستی باید دقیقاً ۱۰ رقم باشد.");
      return;
    }
    if (normalizedEmail && !EMAIL_PATTERN.test(normalizedEmail)) {
      setError("ایمیل را با قالب معتبر وارد کنید.");
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
      setError(err instanceof Error ? err.message : "خطای ناشناخته رخ داد.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {error ? <Alert variant="destructive"><AlertTitle>خطا</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <form onSubmit={submit} className="flex flex-col gap-4">
        <FormSection title="مشخصات" description="نوع پروفایل و شناسه مالی">
          <ProfileTypeField value={profileType} onValueChange={setProfileType} />
          <DisplayNameField id="profile-name" label="نام نمایشی" value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
          <AppTextField
            id="profile-national"
            label={profileType === "legal" ? "شناسه ملی" : "کد ملی"}
            value={nationalID}
            dir="ltr"
            inputMode="numeric"
            maxLength={profileType === "legal" ? 11 : 10}
            onChange={(event) => setNationalID(normalizeDigits(event.target.value).slice(0, profileType === "legal" ? 11 : 10))}
          />
          {profileType === "legal" ? (
            <AppTextField
              id="profile-economic"
              label="کد اقتصادی"
              value={economicCode}
              dir="ltr"
              inputMode="numeric"
              maxLength={12}
              onChange={(event) => setEconomicCode(normalizeDigits(event.target.value).slice(0, 12))}
            />
          ) : null}
        </FormSection>
        <FormSection title="تماس و نشانی" description="برای نمایش در سندهای مالی">
          <AppTextField
            id="profile-postal"
            label="کد پستی"
            value={postalCode}
            dir="ltr"
            inputMode="numeric"
            maxLength={10}
            onChange={(event) => setPostalCode(normalizeDigits(event.target.value).slice(0, 10))}
          />
          <AppTextField id="profile-address" label="نشانی" value={address} onChange={(event) => setAddress(event.target.value)} />
          <AppTextField id="profile-phone" label="تلفن" value={phone} onChange={(event) => setPhone(event.target.value)} />
          <AppTextField
            id="profile-email"
            label="ایمیل"
            value={email}
            type="email"
            dir="ltr"
            inputMode="email"
            autoComplete="email"
            onChange={(event) => setEmail(event.target.value)}
          />
        </FormSection>
        <Button type="submit" disabled={loading} className="self-end">{loading ? "در حال ذخیره..." : "ذخیره پروفایل"}</Button>
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
  const [activeTab, setActiveTab] = useState("invoices");
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
      { accessorKey: "net_adjustment_irr", header: "اصلاحیه (ریال)", cell: ({ row }) => formatAmountValue(row.original.net_adjustment_irr) },
      { accessorKey: "final_amount_irr", header: "مبلغ نهایی (ریال)", cell: ({ row }) => formatAmountValue(row.original.final_amount_irr) },
      { accessorKey: "period_end", header: "پایان دوره", cell: ({ row }) => formatDate(row.original.period_end) },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status)}</StatusBadge> },
      { accessorKey: "invoice_number", header: "شماره" },
    ],
    [],
  );
  const txColumns = useMemo<ColumnDef<WalletTransaction>[]>(
    () => [
      {
        accessorKey: "type",
        header: "نوع",
        cell: ({ row }) => (
          <div className="flex items-center gap-2">
            {transactionTypeIcon(row.original.type)}
            <span>{mapTransactionTypeLabel(row.original.type)}</span>
          </div>
        ),
      },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status)}</StatusBadge> },
      { accessorKey: "amount_irr", header: "مبلغ (ریال)", cell: ({ row }) => formatAmountValue(row.original.amount_irr) },
      { accessorKey: "balance_after_irr", header: "مانده بعد (ریال)", cell: ({ row }) => formatAmountValue(row.original.balance_after_irr) },
      { accessorKey: "created_at", header: "زمان", cell: ({ row }) => formatDate(row.original.created_at) },
    ],
    [],
  );
  const discountColumns = useMemo<ColumnDef<DiscountRedemption>[]>(
    () => [
      { accessorKey: "code", header: "کد" },
      { accessorKey: "discount_type", header: "نوع", cell: ({ row }) => mapDiscountTypeLabel(row.original.discount_type) },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{mapDocumentStatusLabel(row.original.status)}</StatusBadge> },
      { accessorKey: "used_amount_irr", header: "مصرف‌شده (ریال)", cell: ({ row }) => formatAmountValue(row.original.used_amount_irr) },
      { accessorKey: "created_at", header: "ثبت", cell: ({ row }) => formatDate(row.original.created_at) },
    ],
    [],
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
      <Tabs value={activeTab} onValueChange={setActiveTab} className="min-w-0 w-full">
        <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="shrink-0 self-start"
            aria-label="خروجی CSV"
            title="خروجی CSV"
            onClick={() => {
              if (activeTab === "invoices") {
                downloadCSV(
                  ["شماره", "وضعیت", "پایان دوره", "مبلغ نهایی (ریال)", "اصلاحیه (ریال)"],
                  invoiceRows.map((item) => [
                    item.invoice_number,
                    mapDocumentStatusLabel(item.status),
                    formatDate(item.period_end),
                    formatAmountValue(item.final_amount_irr),
                    formatAmountValue(item.net_adjustment_irr),
                  ]),
                  "billing-invoices.csv",
                );
                return;
              }
              if (activeTab === "transactions") {
                downloadCSV(
                  ["نوع", "وضعیت", "مبلغ (ریال)", "مانده بعد (ریال)", "زمان"],
                  transactionRows.map((item) => [
                    mapTransactionTypeLabel(item.type),
                    mapDocumentStatusLabel(item.status),
                    formatAmountValue(item.amount_irr),
                    formatAmountValue(item.balance_after_irr),
                    formatDate(item.created_at),
                  ]),
                  "billing-transactions.csv",
                );
                return;
              }
              downloadCSV(
                ["کد", "نوع", "وضعیت", "مصرف‌شده (ریال)", "ثبت"],
                discountRows.map((item) => [
                  item.code,
                  mapDiscountTypeLabel(item.discount_type),
                  mapDocumentStatusLabel(item.status),
                  formatAmountValue(item.used_amount_irr),
                  formatDate(item.created_at),
                ]),
                "billing-discounts.csv",
              );
            }}
          >
            <Download className="size-4" />
          </Button>
          <TabsList className="w-full justify-start overflow-x-auto sm:w-fit sm:self-end">
            <TabsTrigger value="invoices"><FileText className="me-2 size-4" />فاکتورها</TabsTrigger>
            <TabsTrigger value="transactions"><ReceiptText className="me-2 size-4" />تراکنش‌ها</TabsTrigger>
            <TabsTrigger value="discounts"><Tag className="me-2 size-4" />تخفیف‌ها</TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="invoices">
          <DocumentTableSection
            columns={invoiceColumns}
            rows={invoiceRows}
            emptyLabel="فاکتوری صادر نشده است."
            filters={[
              {
                value: invoiceStatus,
                onValueChange: (value) => {
                  setInvoiceStatus(value);
                  setInvoicePage(1);
                },
                options: buildFilterOptions(invoices.map((item) => item.status), mapDocumentStatusLabel),
                label: "وضعیت",
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
            emptyLabel="تراکنشی ثبت نشده است."
            filters={[
              {
                value: transactionType,
                onValueChange: (value) => {
                  setTransactionType(value);
                  setTransactionPage(1);
                },
                options: buildFilterOptions(transactions.map((item) => item.type), mapTransactionTypeLabel),
                label: "نوع",
              },
              {
                value: transactionStatus,
                onValueChange: (value) => {
                  setTransactionStatus(value);
                  setTransactionPage(1);
                },
                options: buildFilterOptions(transactions.map((item) => item.status), mapDocumentStatusLabel),
                label: "وضعیت",
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
            emptyLabel="تخفیفی استفاده نشده است."
            filters={[
              {
                value: discountType,
                onValueChange: (value) => {
                  setDiscountType(value);
                  setDiscountPage(1);
                },
                options: buildFilterOptions(discounts.map((item) => item.discount_type), mapDiscountTypeLabel),
                label: "نوع",
              },
              {
                value: discountStatus,
                onValueChange: (value) => {
                  setDiscountStatus(value);
                  setDiscountPage(1);
                },
                options: buildFilterOptions(discounts.map((item) => item.status), mapDocumentStatusLabel),
                label: "وضعیت",
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
      <DataTable columns={columns} data={paginatedRows} emptyLabel={emptyLabel} />
      <div className="flex flex-col gap-3 border-t border-border pt-4 text-sm sm:flex-row sm:items-center sm:justify-between">
        <div className="text-muted-foreground">
          نمایش {toPersianNumber(pageStart + 1)} تا {toPersianNumber(pageEnd)} از {toPersianNumber(rows.length)} ردیف
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" disabled={safePage <= 1} onClick={() => onPageChange(safePage - 1)}>
            قبلی
          </Button>
          <span className="min-w-20 text-center text-xs text-muted-foreground">
            صفحه {toPersianNumber(safePage)} از {toPersianNumber(pageCount)}
          </span>
          <Button type="button" variant="outline" size="sm" disabled={safePage >= pageCount} onClick={() => onPageChange(safePage + 1)}>
            بعدی
          </Button>
        </div>
      </div>
    </div>
  );
}

