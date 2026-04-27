import { redirect } from "next/navigation";

import LoginForm from "@/app/components/login-form";
import { getAuthContext } from "@/lib/auth/session";
import { createCaptchaChallenge } from "@/lib/auth/store";

export const dynamic = "force-dynamic";

export default async function LoginPage() {
  const auth = await getAuthContext();
  if (auth) {
    redirect("/");
  }

  const initialCaptcha = await createCaptchaChallenge();
  return <LoginForm initialCaptcha={initialCaptcha} />;
}
