"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { BarChart3, FileImage, FileText, ReceiptText, Tag, Trash2, Upload, WalletCards } from "lucide-react";
import { Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import { AppCard, AppSelectField, FormSection, MetricCard, SettingsCard, StatusBadge } from "@/components/app-ui";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
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
  buildBillingMonthlyPath,
  buildBillingSummaryPath,
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
} from "@/app/billing/_shared";
import { cn } from "@/lib/utils";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
};

const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_RECEIPT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function BankTransferPanel({ initialItems, initialError }: BankTransferPanelProps) {
  const router = useRouter();
  const [amountToman, setAmountToman] = useState("1000000");
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
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge> },
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

      const resp = await fetch("/api/wallet/bank-transfers", {
        method: "POST",
        body: formData,
      });
      const payload = (await resp.json().catch(() => null)) as { error?: { message?: string } } | null;
      if (!resp.ok) throw new Error(payload?.error?.message ?? "ثبت واریز بانکی ناموفق بود.");
      router.refresh();
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
      <SettingsCard title="ثبت واریز بانکی" description="پس از ثبت، تیم پشتیبانی رسید را بررسی و کیف پول را شارژ می‌کند.">
        <form onSubmit={submit} className="grid gap-4 md:grid-cols-2">
          <Field id="bank-amount" label="مبلغ (تومان)" value={amountToman} onChange={setAmountToman} dir="ltr" inputMode="numeric" />
          <Field id="bank-tracking" label="شماره پیگیری" value={trackingNumber} onChange={setTrackingNumber} dir="ltr" />
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
      const [summaryResp, monthlyResp] = await Promise.all([
        fetch(buildBillingSummaryPath(nextPreset)),
        fetch(buildBillingMonthlyPath(nextMonth)),
      ]);
      const [summaryPayload, monthlyPayload] = await Promise.all([summaryResp.json(), monthlyResp.json()]);
      if (!summaryResp.ok) throw new Error(summaryPayload?.error?.message ?? "گزارش خلاصه دریافت نشد.");
      if (!monthlyResp.ok) throw new Error(monthlyPayload?.error?.message ?? "گزارش ماهانه دریافت نشد.");
      setSummary(summaryPayload as BillingUsageSummary);
      setMonthly(monthlyPayload as MonthlyUsageDetail);
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
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <SettingsCard
          title="جزئیات ماهانه"
          description={`تجمیع مصرف، توکن و هزینه نهایی به تفکیک ارائه‌دهنده برای ${formatSelectedJalaliMonth(monthKey)}`}
          className="max-w-none"
        >
          <DataTable columns={columns} data={monthly?.providers ?? []} emptyLabel="داده‌ای برای این ماه وجود ندارد." dir="rtl" />
        </SettingsCard>
        <AppCard title="سهم ارائه‌دهنده‌ها" description="بر اساس هزینه نهایی">
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie data={pieData} dataKey="value" nameKey="name" outerRadius={86} fill="hsl(var(--primary))" />
                <Tooltip formatter={(value) => formatIRR(Number(value))} />
              </PieChart>
            </ResponsiveContainer>
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

type BillingProfileType = "individual" | "legal";

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
      const resp = await fetch("/api/billing/profile", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
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
        }),
      });
      const payload = (await resp.json().catch(() => null)) as { error?: { message?: string } } | null;
      if (!resp.ok) throw new Error(payload?.error?.message ?? "ذخیره پروفایل ناموفق بود.");
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
          <Field id="profile-name" label="نام نمایشی" value={displayName} onChange={setDisplayName} />
          <Field
            id="profile-national"
            label={profileType === "legal" ? "شناسه ملی" : "کد ملی"}
            value={nationalID}
            dir="ltr"
            inputMode="numeric"
            maxLength={profileType === "legal" ? 11 : 10}
            onChange={(value) => setNationalID(normalizeDigits(value).slice(0, profileType === "legal" ? 11 : 10))}
          />
          {profileType === "legal" ? (
            <Field
              id="profile-economic"
              label="کد اقتصادی"
              value={economicCode}
              dir="ltr"
              inputMode="numeric"
              maxLength={12}
              onChange={(value) => setEconomicCode(normalizeDigits(value).slice(0, 12))}
            />
          ) : null}
        </FormSection>
        <FormSection title="تماس و نشانی" description="برای نمایش در سندهای مالی">
          <Field
            id="profile-postal"
            label="کد پستی"
            value={postalCode}
            dir="ltr"
            inputMode="numeric"
            maxLength={10}
            onChange={(value) => setPostalCode(normalizeDigits(value).slice(0, 10))}
          />
          <Field id="profile-address" label="نشانی" value={address} onChange={setAddress} />
          <Field id="profile-phone" label="تلفن" value={phone} onChange={setPhone} />
          <Field
            id="profile-email"
            label="ایمیل"
            value={email}
            type="email"
            dir="ltr"
            inputMode="email"
            autoComplete="email"
            onChange={setEmail}
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
  const invoiceColumns = useMemo<ColumnDef<Invoice>[]>(
    () => [
      { accessorKey: "net_adjustment_irr", header: "اصلاحیه", cell: ({ row }) => formatIRR(row.original.net_adjustment_irr) },
      { accessorKey: "final_amount_irr", header: "مبلغ نهایی", cell: ({ row }) => formatIRR(row.original.final_amount_irr) },
      { accessorKey: "period_end", header: "پایان دوره", cell: ({ row }) => formatDate(row.original.period_end) },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge> },
      { accessorKey: "invoice_number", header: "شماره" },
    ],
    [],
  );
  const txColumns = useMemo<ColumnDef<WalletTransaction>[]>(
    () => [
      { accessorKey: "type", header: "نوع" },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge> },
      { accessorKey: "amount_irr", header: "مبلغ", cell: ({ row }) => formatIRR(row.original.amount_irr) },
      { accessorKey: "balance_after_irr", header: "مانده بعد", cell: ({ row }) => formatIRR(row.original.balance_after_irr) },
      { accessorKey: "created_at", header: "زمان", cell: ({ row }) => formatDate(row.original.created_at) },
    ],
    [],
  );
  const discountColumns = useMemo<ColumnDef<DiscountRedemption>[]>(
    () => [
      { accessorKey: "code", header: "کد" },
      { accessorKey: "discount_type", header: "نوع" },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge> },
      { accessorKey: "used_amount_irr", header: "مصرف‌شده", cell: ({ row }) => formatIRR(row.original.used_amount_irr) },
      { accessorKey: "created_at", header: "ثبت", cell: ({ row }) => formatDate(row.original.created_at) },
    ],
    [],
  );

  return (
    <SettingsCard title="اسناد مالی" description="همه فاکتورها، تراکنش‌ها و تخفیف‌های ارایه‌شده را یکجا ببینید." className="max-w-none">
      <Tabs defaultValue="invoices" className="w-full">
        <TabsList className="self-end" dir="rtl">
          <TabsTrigger value="invoices"><FileText className="me-2 size-4" />فاکتورها</TabsTrigger>
          <TabsTrigger value="transactions"><ReceiptText className="me-2 size-4" />تراکنش‌ها</TabsTrigger>
          <TabsTrigger value="discounts"><Tag className="me-2 size-4" />تخفیف‌ها</TabsTrigger>
        </TabsList>
        <TabsContent value="invoices"><DataTable columns={invoiceColumns} data={invoices} emptyLabel="فاکتوری صادر نشده است." dir="rtl" /></TabsContent>
        <TabsContent value="transactions"><DataTable columns={txColumns} data={transactions} emptyLabel="تراکنشی ثبت نشده است." dir="rtl" /></TabsContent>
        <TabsContent value="discounts"><DataTable columns={discountColumns} data={discounts} emptyLabel="تخفیفی استفاده نشده است." dir="rtl" /></TabsContent>
      </Tabs>
    </SettingsCard>
  );
}

function ProfileTypeField({
  value,
  onValueChange,
}: {
  value: BillingProfileType;
  onValueChange: (value: BillingProfileType) => void;
}) {
  const options = [
    { value: "individual" as const, label: "حقیقی" },
    { value: "legal" as const, label: "حقوقی" },
  ];

  return (
    <div className="flex flex-col gap-2">
      <Label>نوع</Label>
      <RadioGroup
        dir="rtl"
        value={value}
        onValueChange={(nextValue) => onValueChange(nextValue as "individual" | "legal")}
        className="grid grid-cols-2 gap-2"
      >
        {options.map((option) => (
          <Label
            key={option.value}
            htmlFor={`profile-type-${option.value}`}
            className={cn(
              "flex cursor-pointer items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm font-medium transition-colors rtl:flex-row",
              value === option.value && "border-primary bg-accent/40 text-accent-foreground",
            )}
          >
            <span>{option.label}</span>
            <RadioGroupItem id={`profile-type-${option.value}`} value={option.value} />
          </Label>
        ))}
      </RadioGroup>
    </div>
  );
}

function Field({
  id,
  label,
  value,
  onChange,
  type = "text",
  dir,
  inputMode,
  maxLength,
  autoComplete,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  dir?: "rtl" | "ltr" | "auto";
  inputMode?: React.HTMLAttributes<HTMLInputElement>["inputMode"];
  maxLength?: number;
  autoComplete?: string;
}) {
  return (
    <div className="flex flex-col gap-2">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        dir={dir}
        inputMode={inputMode}
        type={type}
        maxLength={maxLength}
        autoComplete={autoComplete}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}

function normalizeDigits(value: string) {
  return value
    .trim()
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)))
    .replace(/\D/g, "");
}

function JalaliDateField({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="flex gap-2">
        <Input
          id={id}
          dir="ltr"
          inputMode="numeric"
          value={value}
          onChange={(e) => onChange(normalizeJalaliDateInput(e.target.value))}
          placeholder="1405-04-17"
        />
        <Button type="button" variant="outline" onClick={() => onChange(getCurrentJalaliDateValue())}>
          امروز
        </Button>
      </div>
      <p className="text-xs leading-5 text-muted-foreground">تاریخ را با فرمت شمسی `YYYY-MM-DD` وارد کنید.</p>
    </div>
  );
}

function formatFileSize(size: number) {
  if (size >= 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }
  if (size >= 1024) {
    return `${Math.round(size / 1024)} KB`;
  }
  return `${size} B`;
}
