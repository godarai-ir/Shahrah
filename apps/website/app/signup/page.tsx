import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { mobileAuthUIEnabled } from "@/app/lib/mobile-auth-ui";
import { buildMetadata } from "@/app/lib/seo";
import { Signup } from "@/app/pages/Signup";

export const metadata: Metadata = buildMetadata({
  title: "Sign Up",
  description:
    "Join GodarAI early access and get started with secure, scalable AI gateway infrastructure.",
  path: "/signup",
});

export default function Page() {
  if (!mobileAuthUIEnabled) {
    redirect("/onboarding");
  }

  return <Signup />;
}
