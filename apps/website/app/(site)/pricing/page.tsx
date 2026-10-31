import type { Metadata } from "next";
import { SeatPricing, type PublicSeatType } from "@/app/pages/SeatPricing";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "قیمت‌گذاری",
  description:
    "صندلی رایگان تا ۵ عضو، و صندلی‌های Pro و Max با قیمت ماهانه برای هر عضو؛ همه‌ی امکانات گدارAI برای همه، با پرداخت پیش از استفاده و بدون تمدید خودکار.",
  path: "/pricing",
});

export const dynamic = "force-dynamic";

export default async function Page() {
  return <SeatPricing seatTypes={await loadSeatTypes()} />;
}

async function loadSeatTypes(): Promise<PublicSeatType[] | null> {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const url = new URL("/api/v1/seat-types", baseURL.endsWith("/") ? baseURL : `${baseURL}/`);
    const response = await fetch(url, { next: { revalidate: 300 } });
    if (!response.ok) return null;
    return ((await response.json()) as { data: PublicSeatType[] }).data;
  } catch {
    return null;
  }
}
