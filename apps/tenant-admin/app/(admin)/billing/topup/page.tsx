import { redirect } from "next/navigation";

// Wallet top-ups ended with the credit cutover; credits are bought as packs.
export default function RetiredWalletPage() {
  redirect("/billing/credits");
}
