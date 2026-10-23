"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import type { ColumnDef } from "@tanstack/react-table";
import { Building2, CreditCard, FileImage, Landmark, Trash2, Upload, type LucideIcon } from "lucide-react";

import { AppTextField } from "@/components/app-ui/AppTextField";
import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { FormSection } from "@/components/app-ui/FormSection";
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
import {
  type BankTransferRequest,
  type BillingProfile,
  getCurrentJalaliDateValue,
  normalizeJalaliDateInput,
  statusTone,
} from "@/features/billing/domain";
import { submitBankTransfer, updateBillingProfile } from "@/features/billing/api";
import {
  formatFileSize,
  JalaliDateField,
  mapDocumentStatusLabel,
  normalizeDigits,
  ProfileTypeField,
  type BillingProfileType,
} from "@/features/billing/billing-ui-helpers";
import { useBillingFormat } from "@/features/billing/use-billing-format";

type BankTransferPanelProps = {
  initialItems: BankTransferRequest[];
  initialError: string | null;
  /** A credit pack or seat order to pay: the amount is fixed to its payable total and the history is hidden. */
  packOrder?: { id: string; payableIRR: number; seat?: boolean };
  onSubmitted?: () => void;
};

const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_RECEIPT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const BANK_TRANSFER_ACCOUNT_INFO = {
  iban: "IR50 0560 9701 8000 0782 3380 01",
  cardNumber: "6219 8610 8689 9761",
} as const;

export function BankTransferPanel({ initialItems, initialError, packOrder, onSubmitted }: BankTransferPanelProps) {
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

    const toman = packOrder ? packOrder.payableIRR / 10 : Number(amountToman.replaceAll(",", ""));
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
      formData.set("amount_irr", String(packOrder ? packOrder.payableIRR : Math.round(toman * 10)));
      if (packOrder) formData.set(packOrder.seat ? "seat_order_id" : "credit_pack_order_id", packOrder.id);
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
      onSubmitted?.();
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
          <AppTextField
            id="bank-amount"
            label={t("bankTransfer.form.amount")}
            value={packOrder ? format.toman(packOrder.payableIRR) : amountToman}
            onChange={(event) => setAmountToman(event.target.value)}
            readOnly={Boolean(packOrder)}
            description={packOrder ? t("bankTransfer.form.packAmountHint") : undefined}
            dir="ltr"
            inputMode="numeric"
          />
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
      {packOrder ? null : (
        <SettingsCard title={t("bankTransfer.list.title")} description={t("bankTransfer.list.description")}>
          <DataTable tableId="billing-bank-transfers" columns={columns} data={initialItems} emptyLabel={t("bankTransfer.list.empty")} />
        </SettingsCard>
      )}
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
