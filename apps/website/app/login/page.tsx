import type { Metadata } from "next";
import { Login } from '@/app/pages/Login';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Login",
  description: "Login page for Shahrah users.",
  path: "/login",
  noIndex: true,
});

export default function Page() {
  return <Login />;
}
