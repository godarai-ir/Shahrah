"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { faIR } from "date-fns/locale";
import { BarChart3, CalendarIcon, FileImage, FileText, ReceiptText, Tag, Trash2, Upload, WalletCards } from "lucide-react";
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
import { Calendar } from "@/components/ui/calendar";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  type BankTransferRequest,
  type BillingProfile,
  type BillingProviderUsage,
  type BillingUsageSummary,
  type DiscountRedemption,
  type Invoice,
  type MonthlyUsageDetail,
  type WalletTransaction,
  formatDate,
  formatDateOnly,
  formatIRR,
  formatToman,
  statusTone,
} from "@/app/billing/_shared";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
};

const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_RECEIPT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];

export function BankTransferPanel({ initialItems, initialError }: BankTransferPanelProps) {
  const router = useRouter();
  const [amountToman, setAmountToman] = useState("1000000");
  const [trackingNumber, setTrackingNumber] = useState("");
  const [bankTransactionDate, setBankTransactionDate] = useState(format(new Date(), "yyyy-MM-dd"));
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
    if (!receiptImage) {
      setError("تصویر رسید را بارگذاری کنید.");
      return;
    }

    setLoading(true);
    try {
      const formData = new FormData();
      formData.set("amount_irr", String(Math.round(toman * 10)));
      formData.set("tracking_number", trackingNumber.trim());
      formData.set("bank_transaction_at", new Date(`${bankTransactionDate}T12:00:00`).toISOString());
      formData.set("receipt_image", receiptImage);

      const resp = await fetch("/api/wallet/bank-transfers", {
        method: "POST",
        body: formData,
      });
      const payload = (await resp.json().catch(() => null)) as { error?: { message?: string } } | null;
      if (!resp.ok) throw new Error(payload?.error?.message ?? "ثبت واریز بانکی ناموفق بود.");
      router.refresh();
      setTrackingNumber("");
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
          <DatePickerField id="bank-date" label="تاریخ واریز" value={bankTransactionDate} onChange={setBankTransactionDate} />
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
  const [range, setRange] = useState(initialSummary?.range ?? "current_month");
  const [month, setMonth] = useState(initialMonthly?.month ?? new Date().toISOString().slice(0, 7));
  const [summary, setSummary] = useState<BillingUsageSummary | null>(initialSummary);
  const [monthly, setMonthly] = useState<MonthlyUsageDetail | null>(initialMonthly);
  const [error, setError] = useState<string | null>(initialError);
  const [loading, setLoading] = useState(false);

  const columns = useMemo<ColumnDef<BillingProviderUsage>[]>(
    () => [
      { accessorKey: "provider_name", header: "Provider" },
      { accessorKey: "request_count", header: "Requests" },
      { accessorKey: "input_tokens", header: "Input tokens" },
      { accessorKey: "output_tokens", header: "Output tokens" },
      { accessorKey: "raw_cost_irr", header: "Raw cost", cell: ({ row }) => formatIRR(row.original.raw_cost_irr) },
      { accessorKey: "discount_irr", header: "Discount", cell: ({ row }) => formatIRR(row.original.discount_irr) },
      { accessorKey: "cache_savings_irr", header: "Cache saving", cell: ({ row }) => formatIRR(row.original.cache_savings_irr) },
      { accessorKey: "final_cost_irr", header: "Final cost", cell: ({ row }) => formatIRR(row.original.final_cost_irr) },
    ],
    [],
  );

  async function refresh(nextRange = range, nextMonth = month) {
    setLoading(true);
    setError(null);
    try {
      const [summaryResp, monthlyResp] = await Promise.all([
        fetch(`/api/billing/reports/summary?range=${encodeURIComponent(nextRange)}`),
        fetch(`/api/billing/reports/monthly?month=${encodeURIComponent(nextMonth)}`),
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
      <div className="grid gap-4 md:grid-cols-4">
        <MetricCard icon={WalletCards} label="مصرف دوره" value={formatIRR(summary?.total_usage_irr)} />
        <MetricCard icon={Tag} label="تخفیف" value={formatIRR(summary?.discount_irr)} />
        <MetricCard icon={BarChart3} label="صرفه‌جویی cache" value={formatIRR(summary?.cache_savings_irr)} />
        <MetricCard icon={ReceiptText} label="اعتبار هدیه" value={formatIRR(summary?.gift_credit_irr)} />
      </div>
      <SettingsCard title="فیلتر گزارش" description="بازه خلاصه و ماه جزئیات را انتخاب کنید.">
        <div className="grid gap-4 md:grid-cols-3">
          <AppSelectField
            label="بازه"
            value={range}
            onValueChange={(value) => {
              setRange(value);
              void refresh(value, month);
            }}
            options={[
              { value: "current_month", label: "ماه جاری" },
              { value: "3m", label: "۳ ماه" },
              { value: "6m", label: "۶ ماه" },
              { value: "12m", label: "۱۲ ماه" },
            ]}
          />
          <Field id="report-month" type="month" label="ماه" value={month} onChange={setMonth} />
          <div className="flex items-end">
            <Button type="button" onClick={() => void refresh()} disabled={loading}>{loading ? "در حال دریافت..." : "به‌روزرسانی"}</Button>
          </div>
        </div>
      </SettingsCard>
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <SettingsCard title="جزئیات ماهانه" description="تجمیع مصرف، توکن و هزینه نهایی به تفکیک provider" className="max-w-none">
          <DataTable columns={columns} data={monthly?.providers ?? []} emptyLabel="داده‌ای برای این ماه وجود ندارد." />
        </SettingsCard>
        <AppCard title="سهم providerها" description="بر اساس هزینه نهایی">
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

export function BillingProfilePanel({ initialProfile, initialError }: ProfilePanelProps) {
  const router = useRouter();
  const data = initialProfile?.data ?? {};
  const [profileType, setProfileType] = useState(initialProfile?.profile_type ?? "individual");
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
    if (!displayName.trim() || !nationalID.trim()) {
      setError("نام نمایشی و شناسه ملی/کد ملی الزامی است.");
      return;
    }
    if (profileType === "legal" && !economicCode.trim()) {
      setError("برای پروفایل حقوقی، کد اقتصادی الزامی است.");
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
            national_id: nationalID.trim(),
            economic_code: economicCode.trim(),
            postal_code: postalCode.trim(),
            address: address.trim(),
            phone: phone.trim(),
            email: email.trim(),
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
    <div className="space-y-6">
      {error ? <Alert variant="destructive"><AlertTitle>خطا</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <SettingsCard title="پروفایل صدور سند" description="اطلاعات این فرم در snapshot فاکتورهای رسمی نگه‌داری می‌شود.">
        <form onSubmit={submit} className="space-y-4">
          <FormSection title="مشخصات" description="نوع پروفایل و شناسه مالی">
            <AppSelectField
              label="نوع"
              value={profileType}
              onValueChange={setProfileType}
              options={[
                { value: "individual", label: "حقیقی" },
                { value: "legal", label: "حقوقی" },
              ]}
            />
            <Field id="profile-name" label="نام نمایشی" value={displayName} onChange={setDisplayName} />
            <Field id="profile-national" label={profileType === "legal" ? "شناسه ملی" : "کد ملی"} value={nationalID} onChange={setNationalID} />
            {profileType === "legal" ? <Field id="profile-economic" label="کد اقتصادی" value={economicCode} onChange={setEconomicCode} /> : null}
          </FormSection>
          <FormSection title="تماس و نشانی" description="برای نمایش در سندهای مالی">
            <Field id="profile-postal" label="کد پستی" value={postalCode} onChange={setPostalCode} />
            <Field id="profile-address" label="نشانی" value={address} onChange={setAddress} />
            <Field id="profile-phone" label="تلفن" value={phone} onChange={setPhone} />
            <Field id="profile-email" label="ایمیل" value={email} onChange={setEmail} />
          </FormSection>
          <Button type="submit" disabled={loading}>{loading ? "در حال ذخیره..." : "ذخیره پروفایل"}</Button>
        </form>
      </SettingsCard>
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
      { accessorKey: "invoice_number", header: "شماره" },
      { accessorKey: "status", header: "وضعیت", cell: ({ row }) => <StatusBadge status={statusTone(row.original.status)}>{row.original.status}</StatusBadge> },
      { accessorKey: "period_end", header: "پایان دوره", cell: ({ row }) => formatDate(row.original.period_end) },
      { accessorKey: "final_amount_irr", header: "مبلغ نهایی", cell: ({ row }) => formatIRR(row.original.final_amount_irr) },
      { accessorKey: "net_adjustment_irr", header: "Adjustment", cell: ({ row }) => formatIRR(row.original.net_adjustment_irr) },
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
    <SettingsCard title="اسناد مالی" description="فاکتورها، ledger کیف پول و تخفیف‌های استفاده‌شده" className="max-w-none">
      <Tabs defaultValue="invoices">
        <TabsList>
          <TabsTrigger value="invoices"><FileText className="me-2 size-4" />فاکتورها</TabsTrigger>
          <TabsTrigger value="transactions"><ReceiptText className="me-2 size-4" />تراکنش‌ها</TabsTrigger>
          <TabsTrigger value="discounts"><Tag className="me-2 size-4" />تخفیف‌ها</TabsTrigger>
        </TabsList>
        <TabsContent value="invoices"><DataTable columns={invoiceColumns} data={invoices} emptyLabel="فاکتوری صادر نشده است." /></TabsContent>
        <TabsContent value="transactions"><DataTable columns={txColumns} data={transactions} emptyLabel="تراکنشی ثبت نشده است." /></TabsContent>
        <TabsContent value="discounts"><DataTable columns={discountColumns} data={discounts} emptyLabel="تخفیفی استفاده نشده است." /></TabsContent>
      </Tabs>
    </SettingsCard>
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
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  dir?: "rtl" | "ltr" | "auto";
  inputMode?: React.HTMLAttributes<HTMLInputElement>["inputMode"];
}) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} dir={dir} inputMode={inputMode} type={type} value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}

function DatePickerField({
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
  const selectedDate = parseISODate(value);

  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            id={id}
            variant="outline"
            type="button"
            className="w-full justify-between text-right font-normal"
          >
            <span className="truncate">
              {selectedDate ? format(selectedDate, "PPP", { locale: faIR }) : "یک تاریخ انتخاب کنید"}
            </span>
            <CalendarIcon data-icon="inline-end" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={selectedDate}
            onSelect={(date) => {
              if (date) {
                onChange(format(date, "yyyy-MM-dd"));
              }
            }}
          />
        </PopoverContent>
      </Popover>
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

function parseISODate(value: string) {
  if (!value) return undefined;
  const [year, month, day] = value.split("-").map(Number);
  if (!year || !month || !day) return undefined;
  return new Date(year, month - 1, day);
}
