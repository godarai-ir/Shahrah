"use client";

import { useTranslations } from "next-intl";

import { FormSheet } from "@/components/app-ui/FormSheet";

import { BankTransferPanel } from "../BillingFinancePanels";
import { useBillingFormat } from "../use-billing-format";
import type { SeatOrder } from "./api";

type SeatTransferSheetProps = {
  /** The seat order awaiting a transfer; null keeps the drawer closed. */
  order: SeatOrder | null;
  onClose: () => void;
  onSubmitted: () => void;
};

/** The drawer that records the bank transfer paying a seat order. */
export function SeatTransferSheet({ order, onClose, onSubmitted }: SeatTransferSheetProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();

  return (
    <FormSheet
      open={order !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      size="lg"
      title={t("transfer.title", { count: order?.quantity ?? 0, seat: order ? t(`seats.${order.seat}`) : "" })}
      description={t("transfer.description", { amount: format.toman(order?.payable_irr ?? 0) })}
      closeLabel={t("transfer.close")}
    >
      {order ? (
        <BankTransferPanel
          key={order.id}
          initialItems={[]}
          initialError={null}
          packOrder={{ id: order.id, payableIRR: order.payable_irr, seat: true }}
          onSubmitted={onSubmitted}
        />
      ) : null}
    </FormSheet>
  );
}
