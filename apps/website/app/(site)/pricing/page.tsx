import type { Metadata } from "next";
import { CommercialPricing, type PublicCommercialPlan } from '@/app/pages/CommercialPricing';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "قیمت‌گذاری",
  description:
    "پلن‌های گُدار AI را برای تیم‌های توسعه، کسب‌وکارهای در حال رشد و سازمان‌ها مقایسه کنید.",
  path: "/pricing",
});

export const dynamic = "force-dynamic";

export default async function Page() {
  return <CommercialPricing plans={await loadCommercialPlans()} />;
}

async function loadCommercialPlans(): Promise<PublicCommercialPlan[]> {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const response = await fetch(new URL("/api/v1/commercial-plans", ensureTrailingSlash(baseURL)), { cache: "no-store" });
    if (!response.ok) return [];
    const payload = (await response.json()) as { data?: PublicCommercialPlan[] };
    return payload.data ?? [];
  } catch {
    return [];
  }
}

function ensureTrailingSlash(value: string) {
  return value.endsWith("/") ? value : `${value}/`;
}
