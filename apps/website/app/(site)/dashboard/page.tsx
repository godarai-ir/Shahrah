import type { Metadata } from "next";
import { Dashboard } from '@/app/pages/Dashboard';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Dashboard Overview",
  description:
    "Explore GodarAI dashboard capabilities for monitoring requests, usage, and AI model performance.",
  path: "/dashboard",
});

export default function Page() {
  return <Dashboard />;
}
