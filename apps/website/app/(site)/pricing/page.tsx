import type { Metadata } from "next";
import { SeatPricing } from "@/app/pages/SeatPricing";
import type { PublicPricingChangelog } from "@/app/pages/PricingChangelog";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "قیمت‌گذاری",
  description:
    "صندلی‌های رایگان، Pro و Max برای اعضای تیم و بسته‌های اعتبار برای استخر سازمان؛ همه‌ی امکانات گدارAI برای همه، با پرداخت پیش از استفاده و بدون تمدید خودکار.",
  path: "/pricing",
});

export const dynamic = "force-dynamic";

export default async function Page() {
  return <SeatPricing creditPriceToman={await loadActiveCreditPrice()} />;
}

// Seat prices follow the active credit list price, published in the public pricing changelog.
async function loadActiveCreditPrice(): Promise<number | null> {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const response = await fetch(new URL("/v1/pricing/changelog", ensureTrailingSlash(baseURL)), { cache: "no-store" });
    if (!response.ok) return null;
    const payload = (await response.json()) as PublicPricingChangelog;
    const active = payload.credit_prices?.find((item) => item.status === "active");
    return active && active.price_toman_per_credit > 0 ? active.price_toman_per_credit : null;
  } catch {
    return null;
  }
}

function ensureTrailingSlash(value: string) {
  return value.endsWith("/") ? value : `${value}/`;
}
