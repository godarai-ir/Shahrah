import type { Metadata } from "next";
import { SiteShell } from "@/app/components/SiteShell";
import { Login } from "@/app/pages/Login";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "ورود",
  description: "صفحه ورود کاربران شاهراه.",
  path: "/login",
  noIndex: true,
});

export default function Page() {
  return (
    <SiteShell>
      <Login />
    </SiteShell>
  );
}
