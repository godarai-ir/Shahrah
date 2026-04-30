import type { Metadata } from "next";
import { Signup } from '@/app/pages/Signup';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Sign Up",
  description:
    "Join Shahrah early access and get started with secure, scalable AI gateway infrastructure.",
  path: "/signup",
});

export default function Page() {
  return <Signup />;
}
