"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { BarChart3, FileText, ReceiptText, Tag, WalletCards } from "lucide-react";
import { Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import { AppCard, AppSelectField, FormSection, MetricCard, SettingsCard, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
  formatIRR,
  formatToman,
  statusTone,
} from "@/app/billing/_shared";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
};

export function BankTransferPanel({ initialItems, initialError }: BankTransferPanelProps) {
  const router = useRouter();
  const [amountToman, setAmountToman] = useState("1000000");
  const [trackingNumber, setTrackingNumber] = useState("");
  const [bankTransactionAt, setBankTransactionAt] = useState(new Date().toISOString().slice(0, 16));
  const [receiptFileURL, setReceiptFileURL] = useState("");
  const [error, setError] = useState<string | null>(initialError);
  const [loading, setLoading] = useState(false);

  const columns = useMemo<ColumnDef<BankTransferRequest>[]>(
    () => [
      { accessorKey: "tracking_number", header: "شماره پیگیری" },
      { accessorKey: "amount_irr", header: "مبلغ", cell: ({ row }) => formatToman(row.original.amount_irr) },
      { accessorKey: "bank_transaction_at", header: "تاریخ واریز", cell: ({ row }) => formatDate(row.original.bank_transaction_at) },
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

    setLoading(true);
    try {
      const resp = await fetch("/api/wallet/bank-transfers", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          amount_irr: Math.round(toman * 10),
          tracking_number: trackingNumber.trim(),
          bank_transaction_at: new Date(bankTransactionAt).toISOString(),
          receipt_file_url: receiptFileURL.trim() || undefined,
        }),
      });
      const payload = (await resp.json().catch(() => null)) as { error?: { message?: string } } | null;
      if (!resp.ok) throw new Error(payload?.error?.message ?? "ثبت واریز بانکی ناموفق بود.");
      router.refresh();
      setTrackingNumber("");
      setReceiptFileURL("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "خطای ناشناخته رخ داد.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>خطا</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <SettingsCard title="ثبت واریز بانکی" description="پس از ثبت، تیم پشتیبانی رسید را بررسی و کیف پول را شارژ می‌کند.">
        <form onSubmit={submit} className="grid gap-4 md:grid-cols-2">
          <Field id="bank-amount" label="مبلغ (تومان)" value={amountToman} onChange={setAmountToman} />
          <Field id="bank-tracking" label="شماره پیگیری" value={trackingNumber} onChange={setTrackingNumber} />
          <Field id="bank-date" label="تاریخ و زمان واریز" type="datetime-local" value={bankTransactionAt} onChange={setBankTransactionAt} />
          <Field id="bank-receipt" label="آدرس رسید" value={receiptFileURL} onChange={setReceiptFileURL} />
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
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
}) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} type={type} value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}
