import type { Metadata } from "next";
import { SeatPricing } from "@/app/pages/SeatPricing";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "قیمت‌گذاری",
  description:
    "صندلی رایگان تا ۵ عضو، Pro ماهی ۲٬۲۵۰٬۰۰۰ تومان و Max ماهی ۱۱٬۲۵۰٬۰۰۰ تومان برای هر عضو؛ همه‌ی امکانات گدارAI برای همه، با پرداخت پیش از استفاده و بدون تمدید خودکار.",
  path: "/pricing",
});

export default function Page() {
  return <SeatPricing />;
}
