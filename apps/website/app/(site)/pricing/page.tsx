import type { Metadata } from "next";
import { Pricing } from '@/app/pages/Pricing';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Pricing",
  description:
    "View Shahrah pricing plans for startups and enterprises with predictable AI gateway costs.",
  path: "/pricing",
});

export default function Page() {
  return <Pricing />;
}
