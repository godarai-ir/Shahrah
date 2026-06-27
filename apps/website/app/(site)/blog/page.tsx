import type { Metadata } from "next";
import { Blog } from '@/app/pages/Blog';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Blog",
  description:
    "Read product updates, engineering insights, and AI gateway best practices from the GodarAI team.",
  path: "/blog",
});

export default function Page() {
  return <Blog />;
}
