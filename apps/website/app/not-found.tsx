import type { Metadata } from "next";
import { NotFound } from '@/app/pages/NotFound';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Page Not Found",
  description: "The requested page could not be found.",
  path: "/404",
  noIndex: true,
});

export default function GlobalNotFound() {
  return <NotFound />;
}
