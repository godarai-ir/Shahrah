import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";
import { saveReceiptImage } from "@/app/api/wallet/bank-transfer-receipts/[fileName]/route";
import { buildBankTransferPayload } from "@/app/billing/_shared";

export const dynamic = "force-dynamic";

function pathWithQuery(request: NextRequest) {
  const query = request.nextUrl.searchParams.toString();
  return `/api/v1/wallet/bank-transfers${query ? `?${query}` : ""}`;
}

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: pathWithQuery(request),
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const contentType = request.headers.get("content-type") ?? "";
  if (contentType.includes("multipart/form-data")) {
    const formData = await request.formData();
    const amountIRR = Number(String(formData.get("amount_irr") ?? "").trim());
    const trackingNumber = String(formData.get("tracking_number") ?? "").trim();
    const bankTransactionDateJalali = String(formData.get("bank_transaction_date_jalali") ?? "").trim();
    const receiptImage = formData.get("receipt_image");

    let receiptFileURL: string | undefined;
    if (receiptImage instanceof File && receiptImage.size > 0) {
      const uploaded = await saveReceiptImage(receiptImage, auth.tenantSlug);
      receiptFileURL = uploaded.url;
    }

    return proxyToControlPlane({
      request,
      method: "POST",
      path: "/api/v1/wallet/bank-transfers",
      body: buildBankTransferPayload({
        amountIRR,
        trackingNumber,
        bankTransactionDateJalali,
        receiptFileURL,
      }),
    });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/wallet/bank-transfers",
    body: await request.json().catch(() => undefined),
  });
}
